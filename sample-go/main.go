package main

import (
	"context"
	"fmt"
	"log"

	vault "github.com/hashicorp/vault/api"
)

func main() {

	config := vault.DefaultConfig()
	config.Address = "http://127.0.0.1:8200"

	// create vault api client
	client, err := vault.NewClient(config)
	if err != nil {
		log.Panicf("Failed to create vault client %s \n", err)
	}

	client.SetToken("root")

	// prepare secret data to store
	secretData := map[string]interface{}{
		"password": "root@123",
		"username": "root",
	}

	// storage keys
	mounthPath := "secret"
	storageKey := "my-secret-data"

	// store secret data to vault
	backgroundCtx := context.Background()
	_, err = client.KVv2(mounthPath).Put(backgroundCtx, storageKey, secretData)
	if err != nil {
		log.Panicf("Failed to store secret data %s", err)
	}

	storedSecret, err := client.KVv2(mounthPath).Get(backgroundCtx, storageKey)
	if err != nil {
		log.Panicf("Failed to get secret %s", err)
	}

	username, isUsernameOk := storedSecret.Data["username"].(string)
	password, isPasswordOk := storedSecret.Data["password"].(string)

	if !isUsernameOk || !isPasswordOk {
		log.Panicf("Failed to get back secret data")
	}

	fmt.Printf("Username: %s \n", username)
	fmt.Printf("Password: %s \n", password)

}
