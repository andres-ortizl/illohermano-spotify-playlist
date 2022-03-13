package spotify

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	openGo "github.com/skratchdot/open-golang/open"
	"golang.org/x/oauth2"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	partialAddItemsUri = "/tracks"
	partialAddCoverUri = "/images"
	createUri          = "https://api.spotify.com/v1/users/%s/playlists"
	playListUri        = "https://api.spotify.com/v1/playlists/%s"
	authorizeUri       = "https://accounts.spotify.com/authorize"
	tokenUri           = "https://accounts.spotify.com/api/token"
	redirectUri        = "http://localhost:8080/callback"
)

const (
	// ScopeImageUpload seeks permission to upload images to Spotify on your behalf.
	ScopeImageUpload = "ugc-image-upload"
	// ScopePlaylistModifyPublic seeks write access
	// to a user's public playlists.
	ScopePlaylistModifyPublic = "playlist-modify-public"
	// ScopePlaylistModifyPrivate seeks write access to

)

type Client struct {
	clientID     string
	clientSecret string
	code         string
	httpClient   *http.Client
	userId       string
}

type RestoredPlayList struct {
	Name   string `json:"name"`
	Tracks struct {
		Items []struct {
			Track struct {
				URI string `json:"uri"`
			} `json:"track"`
		} `json:"items"`
	} `json:"tracks"`
}
type createPlaylistParams struct {
	Name   string `json:"name"`
	Public bool   `json:"public"`
}

type addItemsToPlaylistParams struct {
	Uris []string `json:"uris"`
}

var spotifyConfig = oauth2.Config{
	ClientID:     "",
	ClientSecret: "",
	Scopes:       []string{ScopeImageUpload, ScopePlaylistModifyPublic},
	RedirectURL:  redirectUri,
	// This points to our Authorization Server
	// if our Client ID and Client Secret are valid
	// it will attempt to authorize our user
	Endpoint: oauth2.Endpoint{
		AuthURL:  authorizeUri,
		TokenURL: tokenUri,
	},
}

func New(clientID, clientSecret string, userId string) Client {
	spotifyConfig.ClientID = clientID
	spotifyConfig.ClientSecret = clientSecret
	spot := Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		userId:       userId}
	spot.auth()
	return spot
}

func (spotify *Client) AddItemsToPlaylist(uri *url.URL, uris []string) {

	values := &addItemsToPlaylistParams{
		Uris: uris,
	}
	jsonValue, _ := json.Marshal(values)
	resp, err := spotify.httpClient.Post(
		uri.String()+partialAddItemsUri,
		"application/json",
		bytes.NewBuffer(jsonValue))
	check(err, "")
	log.Info("Items added with status : ", resp.StatusCode)
}

func (spotify *Client) AddCoverToPlaylist(uri *url.URL, imagePath string) {
	image, err := readImage(imagePath)
	imageEncoded := base64.StdEncoding.EncodeToString(image)
	check(err, "")
	req, err := http.NewRequest(
		http.MethodPut,
		uri.String()+partialAddCoverUri,
		strings.NewReader(imageEncoded))
	check(err, "")
	req.Header.Add("Content-Type", "image/jpeg")
	resp, err := spotify.httpClient.Do(req)
	check(err, "")
	log.Info("Cover added with status: ", resp.StatusCode)
}

func (spotify *Client) CreatePlaylist(name string) (*url.URL, error) {
	values := &createPlaylistParams{
		Name:   name,
		Public: true,
	}
	jsonValue, _ := json.Marshal(values)
	resp, err := spotify.httpClient.Post(
		fmt.Sprintf(createUri, spotify.userId),
		"application/json",
		bytes.NewReader(jsonValue))
	log.Info("Response status creating new playlist ", resp.Status)
	check(err, "err")
	location, err := resp.Location()
	check(err, "err")
	log.Info(fmt.Sprintf("New playlist reference for %s : %s", name, location))
	return location, err

}

func (spotify *Client) auth() {
	codeCh, err := startWebServer()
	authURL := spotifyConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	err = openURL(authURL)
	if err != nil {
		log.Fatalf("Unable to open authorization URL in web server: %v", err)
	} else {
		log.Info("This program will resume once authorization has been provided.")
	}

	// Wait for the web server to get the code.
	spotify.code = <-codeCh
	tok, err := spotify.getToken()
	check(err, "Error retrieving token")
	spotify.httpClient = spotifyConfig.Client(context.Background(), tok)
}

func (spotify *Client) getToken() (*oauth2.Token, error) {
	exchange, err := spotifyConfig.Exchange(context.Background(), spotify.code)
	check(err, "")
	return exchange, err
}

func (spotify *Client) GetPlaylist(playListId string) (string, error) {
	resp, err := spotify.httpClient.Get(
		fmt.Sprintf(playListUri, playListId))
	check(err, "err")
	log.Info("Request status pulling playlist information: ", resp.Status)
	body, _ := io.ReadAll(resp.Body)
	return string(body), err
}

func (spotify *Client) RestorePlaylist(playListFile string) {
	var playList RestoredPlayList
	var uriList []string
	jsonFile, err := ioutil.ReadFile(playListFile)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened", playListFile)
	// defer the closing of our jsonFile so that we can parse it later on

	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(jsonFile, &playList)
	if err != nil {
		log.Fatal(err)
	}

	playlist, err := spotify.CreatePlaylist(playList.Name)
	if err != nil {
		return
	}

	for i := 0; i < len(playList.Tracks.Items); i++ {
		uriList = append(uriList, playList.Tracks.Items[i].Track.URI)
	}
	spotify.AddItemsToPlaylist(playlist, uriList)

}

func openURL(url string) error {
	err := openGo.Run(url)
	if err != nil {
		log.Errorf("cannot open URL %s on this platform", url)
		return err
	}

	return err
}

func startWebServer() (codeCh chan string, err error) {
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		return nil, err
	}
	codeCh = make(chan string)
	go func() {
		err := http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			code := r.FormValue("code")
			codeCh <- code // send code to OAuth flow
			err := listener.Close()
			if err != nil {
				return
			}
			w.Header().Set("Content-Type", "text/plain")
		}))
		if err != nil {

		}
	}()

	return codeCh, nil
}

func readImage(imagePath string) ([]byte, error) {
	// Max 256 KB
	img, err := os.ReadFile(imagePath)
	check(err, "Error reading file "+imagePath)
	return img, err
}

func check(e error, msg string) {
	if e != nil {
		log.Debug(msg)
		log.Panic(e)
	}
}
