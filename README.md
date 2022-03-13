# Upload / Save your spotify lists

# Instructions

You'll need to set up the following env vars before running the app

- SPOTIFY_CLIENT_SECRET : Your spotify app client secret, you can set up this
  here https://developer.spotify.com/dashboard/applications
- SPOTIFY_CLIENT_ID : Your spotify app client id, you can set up this
  here https://developer.spotify.com/dashboard/applications
- USER_ID : Your can find your id in the following web page https://www.spotify.com/es/account/overview/

You will also need to setup your Redirect URIs in your spotify developer dashboard. Include the
url `http://localhost:8080/callback`, we're using this to get our auth2 token.

# Usage

- Get play list :

<pre>
    <code>
   go run ./cmd/playlist/main.go -cmd getPlaylist -playlistId "34BqdXQe0WRF9Iub096YJM?si=77ec25b3acd5413a"
    </code>
</pre>

- Create play list :

<pre>
    <code>
   go run ./cmd/playlist/main.go \
  -cmd createPlayList \
  -playListName "a_new_random_playlist" \
  -coverPath "w123.png" \
  -songList "spotify:track:1301WleyT98MSxVHPZCA6M,spotify:track:5nWZYhVUe8R15Di5wJlFli"
    </code>
</pre>

- Restore play list

<pre>
    <code>
    go run ./cmd/playlist/main.go -cmd restorePlayList -playListPath playlist.json
    </code>
</pre>
