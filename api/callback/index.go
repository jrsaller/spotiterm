package callback

import (
    "net/http"
    "github.com/jrsaller/spotiterm/api"
)

func Handler(w http.ResponseWriter, r *http.Request) {
    spotifyauth.CompleteAuthHandler(w, r)
}