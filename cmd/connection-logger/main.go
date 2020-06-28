package main

import (
	"encoding/csv"
	"fmt"
	"github.com/jordan-wright/email"
	"github.com/sparrc/go-ping"
	"math/rand"
	"net/smtp"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var PingAddresses = []string{
	"1.1.1.1",        // cloudflare dns
	"1.0.0.1",        // cloudflare dns
	"8.8.8.8",        // google dns
	"8.8.4.4",        // google dns
	"208.67.222.222", // open dns
	"208.67.220.220", // open dns
	"37.235.1.174",   // FreeDNS
	"37.235.1.177",   // FreeDNS
	"209.244.0.3",    // Level 3
	"209.244.0.4",    // Level 3
	"64.6.64.6",      // Verisign
	"64.6.65.6",      // Verisign
	"198.153.192.1",  // Norton
	"198.153.194.1",  // Norton
	"google.com",
	"amazon.com",
	"facebook.com",
	"netflix.com",
}

func main() {

	pingTicker, speedTicker, emailTicker := createTickers()

	pingsFile, pingsCsv, speedsFile, speedsCsv := createCsvFiles()

	for {
		select {
		case <-pingTicker.C:
			//fmt.Println("pinging")
			if err := recordPing(pingsCsv, pingTicker); err != nil {
				fmt.Println(err.Error())
				//pingTicker.Stop()
				//pingTicker = time.NewTicker(time.Hour * 1000000)
			}
		case <-speedTicker.C:
			//fmt.Println("speed testing")
			if err := recordSpeed(speedsCsv, speedTicker); err != nil {
				fmt.Println(err.Error())
				//speedTicker.Stop()
				//speedTicker = time.NewTicker(time.Hour * 1000000)
			}
		case <-emailTicker.C:
			//fmt.Println("emailing")
			pingsCsv.Flush()
			speedsCsv.Flush()
			if err := emailResults(pingsFile, speedsFile); err != nil {
				fmt.Println("error stopped emailing")
				return // if we can't get the results, it's done, there is no more to do
			}
			os.Remove(pingsFile.Name())
			os.Remove(speedsFile.Name())
			pingsFile, pingsCsv, speedsFile, speedsCsv = createCsvFiles()
		}
	}

}

func recordPing(pingsCsv *csv.Writer, pingTicker *time.Ticker) error {
	address, packetLoss := pingRandom(3)
	err := pingsCsv.Write([]string{time.Now().Format(time.RFC3339), address, fmt.Sprintf("%f", packetLoss)})
	if err != nil {
		return fmt.Errorf("error stopped ping writing: %w", err)
	}
	return nil
}

func recordSpeed(speedsCsv *csv.Writer, speedTicker *time.Ticker) error {
	result, err := exec.Command("./speedtest", "-p", "no", "-f", "csv", "--accept-license").Output()
	if err != nil {
		return fmt.Errorf("command error stopped speedtest: %w", err)
	}

	splitOut := strings.Split(string(result), "\",\"")
	if strings.HasPrefix(splitOut[0], "\"") {
		splitOut[0] = splitOut[0][1:]
	}
	if strings.HasSuffix(splitOut[len(splitOut)-1], "\"") {
		splitOut[len(splitOut)-1] = splitOut[len(splitOut)-1][:len(splitOut[len(splitOut)-1])-1]
	}

	// these speeds are measured in Bytes per second
	downSpeed, err := strconv.ParseFloat(splitOut[5], 64)
	upSpeed, err := strconv.ParseFloat(splitOut[6], 64)

	// 1 Megabit = 125000 Byte
	// we would like to list Megabit per second also, for ease of use, because that is the unit our advertised speeds are in
	splitOut = append(splitOut, fmt.Sprintf("%f", downSpeed/125000), fmt.Sprintf("%f", upSpeed/125000))

	err = speedsCsv.Write(append([]string{time.Now().Format(time.RFC3339)}, splitOut...))
	if err != nil {
		return fmt.Errorf("write error stopped speedtest: %w", err)
	}
	return nil
}

func emailResults(pingsFile *os.File, speedsFile *os.File) error {
	e := email.NewEmail()
	e.From = os.Getenv("EMAIL_USER")
	e.To = []string{os.Getenv("EMAIL_TO")}
	e.Subject = "Speed and Ping reports " + time.Now().Format("2006-01-02")
	e.Text = []byte("Attached you should find your internet speed report")

	_, err := e.AttachFile(pingsFile.Name())
	if err != nil {
		fmt.Print("error attaching pings file:", err)
		return err
	}
	_, err = e.AttachFile(speedsFile.Name())
	if err != nil {
		fmt.Print("error attaching speeds file:", err)
		return err
	}
	err = e.Send("smtp.gmail.com:587", smtp.PlainAuth("", os.Getenv("EMAIL_USER"), os.Getenv("EMAIL_PASS"), "smtp.gmail.com"))
	if err != nil {
		fmt.Print("error sending email:", err)
		return err
	}

	pingsFile.Close()
	speedsFile.Close()

	return nil
}

func createTickers() (*time.Ticker, *time.Ticker, *time.Ticker) {
	var pingTicker *time.Ticker
	if len(os.Getenv("PING_INTERVAL_SEC")) < 1 {
		pingTicker = time.NewTicker(time.Second * 30)
	} else {
		parsedDuration, err := time.ParseDuration(os.Getenv("PING_INTERVAL"))
		if err != nil {
			panic(err)
		} else {
			pingTicker = time.NewTicker(parsedDuration)
		}
	}

	var speedTicker *time.Ticker
	if len(os.Getenv("SPEED_INTERVAL")) < 1 {
		speedTicker = time.NewTicker(time.Minute * 2)
	} else {
		parsedDuration, err := time.ParseDuration(os.Getenv("SPEED_INTERVAL"))
		if err != nil {
			panic(err)
		} else {
			speedTicker = time.NewTicker(parsedDuration)
		}
	}

	var emailTicker *time.Ticker
	if len(os.Getenv("EMAIL_INTERVAL")) < 1 {
		emailTicker = time.NewTicker(time.Minute * 5)
	} else {
		parsedDuration, err := time.ParseDuration(os.Getenv("EMAIL_INTERVAL"))
		if err != nil {
			panic(err)
		} else {
			emailTicker = time.NewTicker(parsedDuration)
		}
	}

	return pingTicker, speedTicker, emailTicker
}

func createCsvFiles() (*os.File, *csv.Writer, *os.File, *csv.Writer) {

	// ping file
	pingFile, err := os.Create("pings-" + time.Now().Format("2006-01-02") + ".csv")
	if err != nil {
		panic(err)
	}
	pingsCsv := csv.NewWriter(pingFile)
	err = pingsCsv.Write([]string{"date", "address", "packet_loss"})
	if err != nil {
		panic(err)
	}

	// speeds file
	speedsFile, err := os.Create("speeds-" + time.Now().Format("2006-01-02") + ".csv")
	if err != nil {
		panic(err)
	}
	speedsCsv := csv.NewWriter(speedsFile)
	err = speedsCsv.Write([]string{"date", "server name", "server id", "latency", "jitter", "packet loss", "download", "upload", "download bytes", "upload bytes", "share url", "down Mb/s", "up Mb/s"})
	if err != nil {
		panic(err)
	}

	return pingFile, pingsCsv, speedsFile, speedsCsv
}

func pingRandom(quantity int) (string, float64) {

	address := PingAddresses[rand.Int()%len(PingAddresses)]

	// select a random address to ping
	pinger, err := ping.NewPinger(address)
	if err != nil {
		fmt.Print("it didn't work")
		panic(err)
	}
	pinger.SetPrivileged(true)
	pinger.Timeout = time.Second * 5
	pinger.Count = quantity
	pinger.Run() // blocks until finished

	return address, pinger.Statistics().PacketLoss
}
