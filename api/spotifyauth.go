// This example demonstrates how to authenticate with Spotify using the authorization code flow.
// In order to run this example yourself, you'll need to:
//
//  1. Register an application at: https://developer.spotify.com/my-applications/
//     - Use "http://localhost:8080/callback" as the redirect URI
//  2. Set the SPOTIFY_ID environment variable to the client ID you got in step 1.
//  3. Set the SPOTIFY_SECRET environment variable to the client secret from step 1.
package spotifyauth

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/zmb3/spotify/v2"
	sa "github.com/zmb3/spotify/v2/auth"
)

const redirectURI = "http://spotiterm.vercel.app/api/callback"

// redirectURI is the OAuth redirect URI for the application.
// You must register an application at Spotify's developer portal
// and enter this value.
var (
	spotifyClient = sa.New(
		sa.WithRedirectURL(redirectURI),
		sa.WithScopes(sa.ScopeUserReadCurrentlyPlaying, sa.ScopeUserReadPlaybackState, sa.ScopeUserModifyPlaybackState),
		sa.WithClientID(os.Getenv("SPOTIFY_ID")),
		sa.WithClientSecret(os.Getenv("SPOTIFY_SECRET")),
	)
	ch    = make(chan *spotify.Client)
	state = "abc123"
)


func Handler(w http.ResponseWriter, r *http.Request) {
	url := spotifyClient.AuthURL(state)
	fmt.Fprintf(w, "Login to Spotify at the following link, if it doesn't automatically open: %s", url)
	// http.Redirect(w, r, url, http.StatusFound)
}

func CompleteAuthHandler(w http.ResponseWriter, r *http.Request) {
	tok, err := spotifyClient.Token(r.Context(), state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		fmt.Println(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	// use the token to get an authenticated client
	client := spotify.New(spotifyClient.Client(r.Context(), tok))
	fmt.Fprintf(w, `
        <html>
            <body>
                <p>Login Completed! You can close this tab now.</p>
                <script type="text/javascript">
                    window.close();
                </script>
            </body>
        </html>
    `)
	ch <- client
}

func Init() *spotify.Client {
	client := <-ch
	return client
}
