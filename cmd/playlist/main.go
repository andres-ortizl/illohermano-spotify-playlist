// +build linux

package main

import (
	"fmt"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"illohermano-spotify-list/pkg/spotify"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	/*playListPath, exist := os.LookupEnv("PLAYLIST_PATH")
	if (!exist) {
		log.Error("PLAYLIST_PATH is not defined")
	}
	*/

	userId, exist := os.LookupEnv("USER_ID")
	if (!exist) {
		log.Error("USER_ID is not defined")
	}

	client_secret, exist := os.LookupEnv("SPOTIFY_SECRET")
	if (!exist) {
		log.Error("SPOTIFY_TOKEN is not defined")
	}

	client_id, exist := os.LookupEnv("CLIENT_ID")
	if (!exist) {
		log.Error("CLIENT_ID is not defined")
	}

	spotifyClient := spotify.New(client_id, client_secret, userId)
	fmt.Println(spotifyClient)
	//u, _ := spotifyClient.CreatePlaylist("TESTON")
	//spotifyClient.AddCoverToPlaylist(u, "/Users/andrew/Code/illohermano-spotify-playlist/data/CWGFD2w.png")

}
