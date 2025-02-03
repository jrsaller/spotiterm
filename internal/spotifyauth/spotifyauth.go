// This example demonstrates how to authenticate with Spotify using the authorization code flow.
// In order to run this example yourself, you'll need to:
//
//  1. Register an application at: https://developer.spotify.com/my-applications/
//     - Use "http://localhost:8080/callback" as the redirect URI
//  2. Set the SPOTIFY_ID environment variable to the client ID you got in step 1.
//  3. Set the SPOTIFY_SECRET environment variable to the client secret from step 1.
package spotifyauth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jrsaller/spotiterm/internal/helpers"
	"github.com/zmb3/spotify/v2"
	sa "github.com/zmb3/spotify/v2/auth"
)

// redirectURI is the OAuth redirect URI for the application.
// You must register an application at Spotify's developer portal
// and enter this value.
const redirectURI = "http://localhost:8080/callback"

var (
	spotifyClient = sa.New(
		sa.WithRedirectURL(redirectURI),
		sa.WithScopes(sa.ScopeUserReadCurrentlyPlaying, sa.ScopeUserReadPlaybackState, sa.ScopeUserModifyPlaybackState),
		sa.WithClientID(os.Getenv("SPOTIFY_ID")),
		sa.WithClientSecret(os.Getenv("SPOTIFY_SECRET")),
	)
	ch = make(chan *spotify.Client)
	state = "abc123"
)

func Init() *spotify.Client {
	// first start an HTTP server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create an HTTP server
	server := &http.Server{Addr: ":8080"}
	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// log.Println("Got request for:", r.URL.String())
	})
	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	url := spotifyClient.AuthURL(state)
	fmt.Println("Login to Spotify at the following link, if it doesn't automatically open:", url)
	helpers.OpenURL(url)

	// // wait for auth to complete
	client := <-ch
	// Shut down the server gracefully
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown Failed:", err)
	}

	// // use the client to make calls that require authorization
	user, err := client.CurrentUser(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You are logged in as:", user.ID)
	fmt.Println("You are logged in as:", user.DisplayName)
	return client
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := spotifyClient.Token(r.Context(), state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
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
