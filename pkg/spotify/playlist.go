package spotify

type Song struct {
}

type Playlist struct {
	songs      []Song
	image   string
	name string
	url string
	userId string
}
