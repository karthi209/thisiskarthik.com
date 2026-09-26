package gamestats

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const SteamEndpoint = "https://api.steampowered.com/IPlayerService/GetOwnedGames/v1/"

type SteamGame struct {
	AppID      int64
	Name       string
	Minutes    int64
	LastPlayed int64
}

// FetchSteamLibrary returns every visible owned game, including unplayed games.
func FetchSteamLibrary(ctx context.Context, client *http.Client, endpoint, key, steamID string) ([]SteamGame, error) {
	if key == "" || steamID == "" {
		return nil, fmt.Errorf("set STEAM_API_KEY and STEAM_ID")
	}
	params, _ := json.Marshal(map[string]any{"steamid": steamID, "include_appinfo": true, "include_played_free_games": true})
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid Steam endpoint")
	}
	q := u.Query()
	q.Set("key", key)
	q.Set("input_json", string(params))
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("could not create Steam request")
	}
	if client == nil {
		client = &http.Client{}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Steam request failed; check network access and credentials")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Steam returned HTTP %d", resp.StatusCode)
	}
	var payload struct {
		Response struct {
			Games []struct {
				AppID      int64  `json:"appid"`
				Name       string `json:"name"`
				Minutes    int64  `json:"playtime_forever"`
				LastPlayed int64  `json:"rtime_last_played"`
			} `json:"games"`
		} `json:"response"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 16<<20)).Decode(&payload); err != nil {
		return nil, fmt.Errorf("invalid Steam response")
	}
	if len(payload.Response.Games) == 0 {
		return nil, fmt.Errorf("Steam returned no visible games; check Game details privacy and account ID")
	}
	games := make([]SteamGame, 0, len(payload.Response.Games))
	for _, game := range payload.Response.Games {
		if game.AppID > 0 && game.Name != "" && game.Minutes >= 0 {
			games = append(games, SteamGame{AppID: game.AppID, Name: game.Name, Minutes: game.Minutes, LastPlayed: game.LastPlayed})
		}
	}
	return games, nil
}

// FetchSteam never returns transport error strings: they can contain the API key URL.
func FetchSteam(ctx context.Context, client *http.Client, endpoint, key, steamID string, ids []int64) (map[int64]Stat, error) {
	if key == "" || steamID == "" {
		return nil, fmt.Errorf("set STEAM_API_KEY and STEAM_ID; existing Steam stats preserved")
	}
	params, _ := json.Marshal(map[string]any{"steamid": steamID, "include_appinfo": false, "include_played_free_games": true, "appids_filter": ids})
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid Steam endpoint")
	}
	q := u.Query()
	q.Set("key", key)
	q.Set("input_json", string(params))
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("could not create Steam request")
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Steam request failed; check network access and credentials")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Steam returned HTTP %d", resp.StatusCode)
	}
	var payload struct {
		Response struct {
			Games []struct {
				AppID      int64  `json:"appid"`
				Minutes    *int64 `json:"playtime_forever"`
				LastPlayed int64  `json:"rtime_last_played"`
			} `json:"games"`
		} `json:"response"`
	}
	if err = json.NewDecoder(io.LimitReader(resp.Body, 8<<20)).Decode(&payload); err != nil {
		return nil, fmt.Errorf("invalid Steam response")
	}
	if len(payload.Response.Games) == 0 {
		return nil, fmt.Errorf("Steam returned no visible games; check Game details privacy and account ID")
	}
	result := map[int64]Stat{}
	for _, g := range payload.Response.Games {
		if g.Minutes == nil || *g.Minutes < 0 {
			continue
		}
		result[g.AppID] = Stat{Provider: "steam", Hours: Hours(float64(*g.Minutes) / 60), LastPlayed: timestamp(g.LastPlayed)}
	}
	return result, nil
}
