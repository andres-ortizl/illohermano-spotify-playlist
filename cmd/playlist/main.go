//go:build linux
// +build linux

package main

import (
	"flag"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"illohermano-spotify-list/pkg/spotify"
	"os"
	"strings"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	userId, exist := os.LookupEnv("USER_ID")
	if !exist {
		log.Error("USER_ID is not defined")
	}

	clientSecret, exist := os.LookupEnv("SPOTIFY_CLIENT_SECRET")
	if !exist {
		log.Error("SPOTIFY_CLIENT_SECRET is not defined")
	}

	clientId, exist := os.LookupEnv("SPOTIFY_CLIENT_ID")
	if !exist {
		log.Error("SPOTIFY_CLIENT_ID is not defined")
	}

	spotifyClient := spotify.New(clientId, clientSecret, userId)

	cmd := flag.String("cmd", "", "Cmd Type (getPlayList, createPlayList, restorePlayList)")
	playListId := flag.String("playListId", "", "Id of the playlist you want to download")
	playListName := flag.String("playListName", "", "Name of  the playlist you want to upload")
	coverPath := flag.String("coverPath", "", "Absolute path of your playlist image")
	playListPath := flag.String("playListPath", "", "Absolute path of your playlist json file to restore")
	songList := flag.String("songList", "", "List of songs, comma separated values. Ex : spotify:track:1301WleyT98MSxVHPZCA6M")
	flag.Parse()
	if *cmd == "getPlayList" {
		e, _ := spotifyClient.GetPlaylist(*playListId)
		fileName := "playlist.json"
		f, err := os.Create(fileName)

		if err != nil {
			return
		}
		_, err = f.WriteString(e)
		if err != nil {
			return
		}
		log.Info("Successfully saved playlist in ", fileName)
	} else if *cmd == "createPlayList" {
		u, _ := spotifyClient.CreatePlaylist(*playListName)
		if len(*coverPath) > 0 {
			spotifyClient.AddCoverToPlaylist(u, *coverPath)
		}
		items := strings.Split(*songList, ",")
		spotifyClient.AddItemsToPlaylist(u, items)

	} else if *cmd == "restorePlayList" {
		spotifyClient.RestorePlaylist(*playListPath)
	}

}
