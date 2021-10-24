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
	log.Info("Items added with status: ", resp.StatusCode)
}

func (spotify *Client) CreatePlaylist(name string) (*url.URL, error) {
	values := &createPlaylistParams{
		Name:   name,
		Public: true,
	}
	jsonValue, _ := json.Marshal(values)
	fmt.Println(jsonValue)
	resp, err := spotify.httpClient.Post(
		fmt.Sprintf(createUri, spotify.userId),
		"application/json",
		bytes.NewReader(jsonValue))
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
