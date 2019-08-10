package main

import (
	"github.com/adrianmo/go-nmea"
	"go.bug.st/serial.v1"
	"log"
	"strconv"
	"strings"
	"time"
)

type GpsData struct {
	Timestamp string
	Latitude string
	Longitute string
}

func InitGPS(portDevice string) (serial.Port) {

	mode := &serial.Mode{
		BaudRate: 115200,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open("/dev/" + portDevice, mode)
	if err != nil {
		log.Fatal(err)
	}

	_, err = port.Write([]byte("AT+CGNSPWR=1\r\nAT+CGNSSEQ=\"RMC\"\r\nAT+CGNSINF\r\nAT+CGNSURC=2\r\nAT+CGNSTST=1\r\n"))

	if err != nil {
		log.Fatal(err)
	}

	return port
}

func ReadGPS(port serial.Port) string {
	buff := make([]byte, 1)
	var ret strings.Builder

	for string(buff[0]) != "\n" {
	_, err := port.Read(buff)
	if err != nil {
		log.Printf("Couldn't read GPS coords")
	}
		ret.WriteString(string(buff[0]))
	}
	return ret.String()
}

func writeGpsData(input string) {
	lines := strings.Split(input, "\n")

	for _, line := range lines {
		if strings.Contains(line, "GNRMC") {
				line = strings.TrimSuffix(line, "\n")
				s, err := nmea.Parse(line)
				if err != nil {
					log.Printf("Couldn't parse GPS data: %v", err)
					continue
				}
				if s.DataType() == nmea.TypeRMC {
					m := s.(nmea.RMC)
					now := time.Now()
					GPSdata.Timestamp = strconv.FormatInt(now.Unix(), 10)
					GPSdata.Latitude = convertDMStoDec(nmea.FormatGPS(m.Latitude))
					GPSdata.Longitute = convertDMStoDec(nmea.FormatGPS(m.Longitude))
				}
		} else {
			continue
		}
	}
}

func convertDMStoDec(data string) string {
	var ret strings.Builder
	tmp := strings.Split(data, ".")
	tmp1 := tmp[0]
	secDec := tmp[1]
	minDeg := tmp1[len(tmp1)-2:]
	degDec := tmp1[:len(tmp1)-2]

	minutes, _ := strconv.Atoi(minDeg)
	minutes *= 100
	minutes = minutes / 60

	ret.WriteString(degDec + "." + strconv.Itoa(minutes) + secDec)
	return ret.String()
}

func gpsScanner(stop chan bool) {
	stopscanner := false

	log.Println("Scanner: starting")

	for {
		data := ReadGPS(port)

		writeGpsData(data)

		select {
		case stopscanner = <- stop:
			if stopscanner {
				log.Println("GPS Scanner: stopping")
				port.Close()
				break
			}
		default:
			continue
		}
	}
}