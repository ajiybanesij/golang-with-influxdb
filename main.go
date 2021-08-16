package main

import (
	"fmt"
	_ "github.com/influxdata/influxdb1-client"
	client "github.com/influxdata/influxdb1-client/v2"
	"log"
	"math"
	"time"
)

// Öncelikle sabitleri tanımlanması gerekiyor.
// databasename, username ve password inlfuxdb için gereklidir.
// Hot limit ortamın ulaşabileceği maksimum sıcaklıktır ve bu değere ulaşıldığında klima açılır.
// Cold Limit ortamın ulaşabileceği minimun sıcaklıktır ve bu değere ulaşıldığında klima kapanır .
// deltateml sıcaklık değişimi için gerekli katsayıdır.
// thresold ise limitlerin eşiğidir.

const (
	databaseaddr="http://localhost:8086"
	database = "example"
	username = "root"
	password = "root"
	hotlimit = 45.0
	coldlimit= 10.0
	deltatemp=0.1
	threshold=0.05
)

func influxDBClient() client.Client {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:databaseaddr     ,
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatalln("Error: ", err)
	}
	return c
}

func insertMetrics(c client.Client,temp float64,airConditionerStatus int) {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  database,
		Precision: "s",
	})

	if err != nil {
		log.Fatalln("Error: ", err)
	}

	tags := map[string]string{
		"airConditioner": "Air Conditioner 1",
	}
	fields := map[string]interface{}{
		"degree":  temp,
		"airConditionerStatus":airConditionerStatus,
	}
	point, err := client.NewPoint(
		"temperature",
		tags,
		fields,
		time.Now(),
	)
	if err != nil {
		log.Fatalln("Error: ", err)
	}

	bp.AddPoint(point)

	err = c.Write(bp)
	if err != nil {
		log.Fatal(err)
	}
}

func changeEnvironmentDegree(currentTemp float64, deltaTemp float64,limit float64) float64  {
	return limit + (currentTemp-limit)*math.Exp(-0.300*deltaTemp)
}

func main() {
	c := influxDBClient()

	limit := coldlimit
	var airConditionerStatus =1
	currentTemp := 28.0

	for {
		currentTemp=changeEnvironmentDegree(currentTemp,deltatemp,limit)

		if currentTemp >= hotlimit-threshold {
			airConditionerStatus=1
			limit=coldlimit
		} else if currentTemp <= coldlimit+threshold {
			airConditionerStatus=0
			limit=hotlimit
		}
		fmt.Println(airConditionerStatus,currentTemp)
		insertMetrics(c,currentTemp,airConditionerStatus)
		time.Sleep(100 * time.Millisecond)
	}
}

