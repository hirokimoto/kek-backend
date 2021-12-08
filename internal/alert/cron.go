package alert

import (
	"log"

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
		go sendMessage()
	})
	c.Start()
}
