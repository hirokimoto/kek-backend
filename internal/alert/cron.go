package alert

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"kek-backend/internal/uniswap"

	"github.com/appleboy/go-fcm"
	"github.com/robfig/cron/v3"
)

func sendMessage() {
	// Create the message to be sent.
	msg := &fcm.Message{
		To: "eqGLNjY2A6XP3aEN254CRv:APA91bEqWb7K4lj9snQ9RArYx__KELSqKhIlmKlBTho2u1rFXN2QbJ1vJLXHtEYjVVcqz-DREV8fGMEb6wGhD_HXulURUmoe7CG7Ktk1btaMnFeGx1_5SCRMEIovekmR5PVoP5T-fDBY",
		Data: map[string]interface{}{
			"foo": "bar",
		},
		Notification: &fcm.Notification{
			Title: "Hiroki's Go is awesome",
			Body:  "Hiroki sends a push notification via go backend!",
		},
	}

	// Create a FCM client to send the message.
	client, err := fcm.NewClient("AAAAlSnRveU:APA91bF_XWeThMJnZuUGUyQ5wIBBBRyqGfryJ818ItRFUcJg0HubP6ekcNw0FF-ebQMHFZwva2wfEBIViv9qTh7QTeafiyHk8BWPgdE-j3DQEe2orVHpyayxF7DOyOujlarj2_SEhIr_")
	if err != nil {
		log.Fatalln(err)
	}

	// Send the message and receive the response without retries.
	response, err := client.Send(msg)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("%#v\n", response)
}

func StartCron() {
	c := cron.New(cron.WithSeconds())
	c.AddFunc("@every 5s", func() {
		c1 := make(chan string)
		c2 := make(chan string)

		var ethPrice float64
		var tokenPrice float64
		query1 := uniswap.QueryBundles()
		go uniswap.Request(query1, c1)
		query2 := uniswap.QuertyToken("0x2f02be0c4021022b59e9436f335d69df95e5222a")
		go uniswap.Request(query2, c2)

		select {
		case msg1 := <-c1:
			var bundles uniswap.Bundles
			json.Unmarshal([]byte(msg1), &bundles)
			ethPrice, _ = strconv.ParseFloat(bundles.Data.Bundles[0].EthPrice, 64)
			fmt.Println("TK1: ", ethPrice, tokenPrice)
		case msg2 := <-c2:
			var tokens uniswap.Tokens
			json.Unmarshal([]byte(msg2), &tokens)
			tokenPrice, _ = strconv.ParseFloat(tokens.Data.Tokens[0].DerivedETH, 64)
			fmt.Println("TK2: ", ethPrice, tokenPrice)
		}

		fmt.Println("$$$$$:   ", ethPrice, tokenPrice)
		// go sendMessage()
	})
	c.Start()
}
