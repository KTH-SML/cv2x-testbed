package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

var testCasesFlag = flag.String("cases", "", "Which test cases should be run?")
var testRatesFlag = flag.String("rates", "", "Which rates should be used? (maps to cases)")
var testSizesFlag = flag.String("sizes", "", "Which sizes should be used? (maps to cases)")
var testDurationFlag = flag.Duration("duration", 2*time.Minute, "How long should each test run?")
var testCooldownFlag = flag.Duration("cooldown", 30*time.Second, "How long should the coordinator wait between tests?")
var testRerunsFlag = flag.Uint("times", 1, "How many times should the test suite be run?")
var switchManuallyFlag = flag.Bool("manual", false, "Should the tests be switched manually?")
var enableVerboseFlag = flag.Bool("verbose", false, "Print information about the test cases")

func coordinator() func(*Node) {

	startTime := time.Now().Format("060102_1504")

	testCases := []int{}
	if len(*testCasesFlag) != 0 {
		for _, s := range strings.Split(*testCasesFlag, ",") {
			n, err := strconv.Atoi(s)
			if err != nil {
				panic(err)
			}
			testCases = append(testCases, n)
		}
	}

	testRates := []int{} // [Hz]
	if len(*testRatesFlag) != 0 {
		for _, s := range strings.Split(*testRatesFlag, ",") {
			n, err := strconv.Atoi(s)
			if err != nil {
				panic(err)
			}
			testRates = append(testRates, n)
		}
	}

	testSizes := []int{} // [B]
	if len(*testSizesFlag) != 0 {
		for _, s := range strings.Split(*testSizesFlag, ",") {
			n, err := strconv.Atoi(s)
			if err != nil {
				panic(err)
			}
			testSizes = append(testSizes, n)
		}
	}

	testLoads := []int{} // [%]
	testMobility := []bool{}
	testFeatures := []string{}

	for _, testCase := range testCases {
		if len(*testCasesFlag) != 0 && len(*testRatesFlag) == 0 && len(*testSizesFlag) == 0 {
			// Evaluate digit TCxxx[x]
			switch (testCase / 1) % 10 {
			case 0:
				testRates = append(testRates, 10)
				testSizes = append(testSizes, 1_000)
			case 1:
				testRates = append(testRates, 10)
				testSizes = append(testSizes, 10_000)
			case 2:
				testRates = append(testRates, 20)
				testSizes = append(testSizes, 1_000)
			case 3:
				testRates = append(testRates, 20)
				testSizes = append(testSizes, 10_000)
			case 4:
				testRates = append(testRates, 20)
				testSizes = append(testSizes, 100_000)
			case 5:
				testRates = append(testRates, 10)
				testSizes = append(testSizes, 100_000)
			case 6:
				testRates = append(testRates, 10)
				testSizes = append(testSizes, 500_000)
			default:
				panic(fmt.Sprintf("Unrecognized digit TCxxx[x] (TC%d)", testCase))
			}
		}

		// Evaluate digits TCx[xx]x
		switch (testCase / 10) % 100 {
		case 0:
			testLoads = append(testLoads, 0)
			testMobility = append(testMobility, false)
		case 1:
			testLoads = append(testLoads, 0)
			testMobility = append(testMobility, false)
		case 2:
			testLoads = append(testLoads, 50)
			testMobility = append(testMobility, false)
		case 3:
			testLoads = append(testLoads, 90)
			testMobility = append(testMobility, false)
		case 4:
			testLoads = append(testLoads, 0)
			testMobility = append(testMobility, true)
		case 5:
			testLoads = append(testLoads, 50)
			testMobility = append(testMobility, true)
		case 6:
			testLoads = append(testLoads, 90)
			testMobility = append(testMobility, true)
		case 7:
			testLoads = append(testLoads, 00)
			testMobility = append(testMobility, true)
		case 8:
			testLoads = append(testLoads, 90)
			testMobility = append(testMobility, true)
		case 9:
			testLoads = append(testLoads, 50)
			testMobility = append(testMobility, true)
		case 10:
			testLoads = append(testLoads, 100)
			testMobility = append(testMobility, false)
		case 11:
			testLoads = append(testLoads, 200)
			testMobility = append(testMobility, false)
		case 12:
			testLoads = append(testLoads, 400)
			testMobility = append(testMobility, false)
		case 13:
			testLoads = append(testLoads, 800)
			testMobility = append(testMobility, false)
		case 20:
			testLoads = append(testLoads, 100)
			testMobility = append(testMobility, false)
		case 21:
			testLoads = append(testLoads, 200)
			testMobility = append(testMobility, false)
		case 22:
			testLoads = append(testLoads, 400)
			testMobility = append(testMobility, false)
		case 23:
			testLoads = append(testLoads, 800)
			testMobility = append(testMobility, false)
		case 31:
			testLoads = append(testLoads, 0)
			testMobility = append(testMobility, false)
		case 32:
			testLoads = append(testLoads, 50)
			testMobility = append(testMobility, false)
		case 33:
			testLoads = append(testLoads, 90)
			testMobility = append(testMobility, false)
		default:
			panic(fmt.Sprintf("Unrecognized digits TCx[xx]x (TC%d)", testCase))
		}

		// Evaluate digit TC[x]xxx
		switch (testCase / 1000) % 10 {
		case 0:
			testFeatures = append(testFeatures, "Debug")
		case 1:
			testFeatures = append(testFeatures, "Baseline")
		case 2:
			testFeatures = append(testFeatures, "Absolute Priority")
		default:
			panic(fmt.Sprintf("Unrecognized digit TC[x]xxx (TC%d)", testCase))
		}
	}

	if len(testCases) == 0 {
		panic("No test cases specified")
	}

	if len(testRates) != len(testCases) {
		panic(fmt.Sprintf("Number of test cases (%d) and rates (%d) do not match", len(testCases), len(testRates)))
	}

	if len(testSizes) != len(testCases) {
		panic(fmt.Sprintf("Number of test cases (%d) and sizes (%d) do not match", len(testCases), len(testSizes)))
	}

	testDuration := *testDurationFlag
	testCooldown := *testCooldownFlag
	testReruns := *testRerunsFlag
	switchManually := *switchManuallyFlag
	enableVerbose := *enableVerboseFlag

	numRuns := len(testCases) * int(testReruns)

	err := os.MkdirAll("logs", os.ModePerm)
	if err != nil {
		panic(err)
	}

	return func(node *Node) {
		// clean log at vehicle
		node.remote_set_log("vehicle", []Packet{})

		checkConnection(node)
		time.Sleep(1 * time.Second)

		logDir := path.Join("logs", startTime)
		os.MkdirAll(logDir, os.ModePerm)
		flagsFile, err := os.Create(path.Join(logDir, "flags.yml"))
		if err != nil {
			panic(err)
		}
		defer flagsFile.Close()
		flagsFile.WriteString(fmt.Sprintf("# Test suite started at %s with following test(s): %s\n", startTime, *testCasesFlag))

		fmt.Printf("Starting tests! They will be stored here: %s\n", logDir)

		progress := 0
		for n := 0; n < int(testReruns); n++ {
			for i := 0; i < len(testCases); i++ {
				if switchManually {
					fmt.Printf("Press enter to continue with next case: TC%d", testCases[i])
					fmt.Scanln()
					fmt.Println("")
				} else if progress != 0 {
					time.Sleep(testCooldown)
				}

				timeNow := time.Now().Format("060102_1504")
				fileName := fmt.Sprintf("%s__TC%d.csv", timeNow, testCases[i])
				filePath := path.Join(logDir, fileName)

				if enableVerbose {
					offset := node.ntpClient.Resp.ClockOffset.Milliseconds()
					fmt.Printf("(%d/%d) Running TC%d - NTP offset %d ms\r", progress, numRuns, testCases[i], offset)
				} else {
					fmt.Printf("(%d/%d) Running TC%d\r", progress, numRuns, testCases[i])
				}

				// Set the test case configuration
				node.remote_set("sensor", "rate", testRates[i])
				node.remote_set("sensor", "DATA_SIZE", testSizes[i])
				node.remote_set("sensor", "DATA_SEQ", 0)
				node.remote_set("server", "COMPUTE_TIME", 0)

				// Run the actual test
				runTest(node, testDuration)

				// Retreive the log and save it
				log, err := node.remote_get_log("vehicle")
				if err != nil {
					panic(err)
				}
				node.remote_set_log("vehicle", []Packet{})
				save(log, filePath)

				// Write flags to file
				flagsFile.WriteString(fmt.Sprintf("- case: %d\n", testCases[i]))
				flagsFile.WriteString(fmt.Sprintf("  rate: %d\n", testRates[i]))
				flagsFile.WriteString(fmt.Sprintf("  size: %d\n", testSizes[i]))
				flagsFile.WriteString(fmt.Sprintf("  load: %d\n", testLoads[i]))
				flagsFile.WriteString(fmt.Sprintf("  mobility: %t\n", testMobility[i]))
				flagsFile.WriteString(fmt.Sprintf("  features: \"%s\"\n", testFeatures[i]))
				flagsFile.WriteString(fmt.Sprintf("  datetime: \"%s\"\n", timeNow))
				flagsFile.WriteString(fmt.Sprintf("  duration: %f\n", testDuration.Seconds()))
				flagsFile.WriteString(fmt.Sprintf("  cooldown: %f\n", testCooldown.Seconds()))
				flagsFile.WriteString(fmt.Sprintf("  filename: %s\n", fileName))

				progress++
			}
		}

		// Pause the testing (cause main to stop running)
		fmt.Printf("(%d/%d)\nTests finished!", progress, numRuns)
		node.attr["paused"] = 1
	}
}

func checkConnection(node *Node) {
	runTest(node, 5*time.Second)
	log, err := node.remote_get_log("vehicle")
	if err != nil {
		panic(err)
	}
	if len(log) == 0 {
		panic("Data is not coming through!")
	}
	node.remote_set_log("vehicle", []Packet{})
}

func runTest(n *Node, test_duration time.Duration) {
	n.unpause("vehicle", "server", "sensor")
	time.Sleep(test_duration)
	n.pause("server", "sensor")
	time.Sleep(time.Duration(1))
	n.pause("vehicle") // to give enough time for the vehicle to send the trailing messages
}

func save(log []Packet, filename string) {

	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	csvwrite := csv.NewWriter(file)
	defer csvwrite.Flush()

	csvwrite.Write([]string{"t1", "t2", "t3", "t4", "e1", "e2", "e3", "e4", "x", "y", "yaw", "vel", "lat", "lon", "seq", "valid", "frame_id"})
	for _, packet := range log {
		t1 := strconv.FormatInt(packet.T1, 10)
		t2 := strconv.FormatInt(packet.T2, 10)
		t3 := strconv.FormatInt(packet.T3, 10)
		t4 := strconv.FormatInt(packet.T4, 10)
		e1 := strconv.FormatInt(packet.E1, 10)
		e2 := strconv.FormatInt(packet.E2, 10)
		e3 := strconv.FormatInt(packet.E3, 10)
		e4 := strconv.FormatInt(packet.E4, 10)
		x := strconv.FormatFloat(packet.X, 'f', -1, 64)
		y := strconv.FormatFloat(packet.Y, 'f', -1, 64)
		yaw := strconv.FormatFloat(packet.Yaw, 'f', -1, 64)
		vel := strconv.FormatFloat(float64(packet.V), 'f', -1, 64)
		lat := strconv.FormatFloat(packet.Latitude, 'f', -1, 64)
		lon := strconv.FormatFloat(packet.Longitude, 'f', -1, 64)
		seq := strconv.FormatInt(packet.Header.Seq, 10)
		chk := strconv.Itoa(packet.Chk)
		frame_id := packet.Header.FrameID

		csvwrite.Write([]string{t1, t2, t3, t4, e1, e2, e3, e4, x, y, yaw, vel, lat, lon, seq, chk, frame_id})
	}
}
