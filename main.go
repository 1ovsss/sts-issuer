package main

import (
	"log"
	"os"

	"sts-issuer/internal/api"
	"sts-issuer/internal/notify"
	"sts-issuer/internal/sts"
)

func main() {
	// init aws sdk
	errI := sts.InitCfg()
	if errI != nil {
		log.Fatalf("Failed to init STS client: %v", errI)
	}

	// determine server or cron to run
	id := os.Getenv("STS_CRON_IDENTIFIER")
	if id == "" {
		api.Start()
	} else {
		stsCreds, errS := sts.GetCreds(id)
		if errS != nil {
			log.Fatalf("Failed to get sts creds: %v", errS)
		}
		errR := notify.SendRocketChatNotification(stsCreds, id)
		if errR != nil {
			log.Fatalf("Failed to send Rocket notification: %v", errR)
		}
	}
}
