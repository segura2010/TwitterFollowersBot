package main

import (
    "log"
    "fmt"
    "encoding/json"
    "io/ioutil"
    "time"
    "net/url"
    "net/http"
    //"strings"
    "strconv"
    "bytes"

    "flag"

    "github.com/kurrik/oauth1a"
    "github.com/kurrik/twittergo"
)

// MyConfig struct: for the user config
// This is the struct that the config.json must have
type MyConfig struct {
    ConsumerKey string
    ConsumerSecret string

    AccessToken string
    AccessSecret string

    RefreshTime int // time in minutes to check for new followers and unfollowers
    Username string // username of the user to monitor
}

type FollowersList struct {
    Ids []int64
}

var (
    client *twittergo.Client
    config MyConfig
)

func LoadConfig(filename string) (MyConfig, error){
    var s MyConfig

    bytes, err := ioutil.ReadFile(filename)
    if err != nil {
        return s, err
    }
    // Unmarshal json
    err = json.Unmarshal(bytes, &s)
    return s, err
}

func checkRateLimits(resp *twittergo.APIResponse){
    log.Printf("Rate Limit: %d/%d, Rate Limit Reset: %d (%s)", resp.RateLimitRemaining(), resp.RateLimit(), resp.RateLimitReset().Unix(), resp.RateLimitReset().Format(time.RFC1123))
}

func sendDM(username, text string){
    body := []byte(`{"event": {"type": "message_create", "message_create": {"target": {"recipient_id":"`+ username +`"}, "message_data": {"text": "`+ text +`"}}}}`)

    url := fmt.Sprintf("/1.1/direct_messages/events/new.json")
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
    if err != nil {
        log.Printf("Error sending DM: %s", err)
        return
    }
    req.Header.Set("Content-Type", "application/json")
    _, err = client.SendRequest(req)

    if err != nil {
        log.Printf("Error sending DM: %s", err)
    }
}

func getFollowers(username string) (*FollowersList, error){
    query := url.Values{}
    query.Set("screen_name", username)

    url := fmt.Sprintf("/1.1/followers/ids.json?%v", query.Encode())
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    resp, err := client.SendRequest(req)

    if err != nil {
        return nil, err
    }

    checkRateLimits(resp)

    results := &FollowersList{}
    err = resp.Parse(results)
    if err != nil {
        return nil, err
    }

    return results, nil
}

func getUserInfo(userId int64) (*twittergo.User, error){
    query := url.Values{}
    query.Set("user_id", strconv.FormatInt(userId, 10))

    url := fmt.Sprintf("/1.1/users/show.json?%v", query.Encode())
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    resp, err := client.SendRequest(req)

    if err != nil {
        return nil, err
    }

    results := &twittergo.User{}
    err = resp.Parse(results)
    if err != nil {
        return nil, err
    }

    return results, nil
}

func getUserInfoByUsername(screenname string) (*twittergo.User, error){
    query := url.Values{}
    query.Set("screen_name", screenname)

    url := fmt.Sprintf("/1.1/users/show.json?%v", query.Encode())
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    resp, err := client.SendRequest(req)

    if err != nil {
        return nil, err
    }

    results := &twittergo.User{}
    err = resp.Parse(results)
    if err != nil {
        return nil, err
    }

    return results, nil
}

func IndexOf(array []int64, e int64) (int){
    for i, ele := range array{
        if ele == e{
            return i
        }
    }
    return -1
}

func mainLoop(refreshTime int, username string){
    usrInfo, err := getUserInfoByUsername(username)
    if err != nil {
        log.Printf("Error getting userId!: %s", err)
        return
    }

    log.Printf("Got User ID: %s", usrInfo.IdStr())    

    previousFollowers, err := getFollowers(username)
    if err != nil {
        log.Printf("Error getting initial followers!!: %s", err)
        return
    }

    log.Printf("Initial followers updated!")
    sendDM(usrInfo.IdStr(), "🤖 Started!")

    for{
        // sleep
        time.Sleep(time.Duration(refreshTime) * time.Minute)

        actualFollowers, err := getFollowers(username)
        if err != nil {
            log.Printf("Error getting followers: %s", err)
            sendDM(username, fmt.Sprintf("🤖 Error getting followers: %s", err))
        } else {
            // Check unfollows
            for _, f := range previousFollowers.Ids{
                if IndexOf(actualFollowers.Ids, f) == -1{
                    // Not found -> Unfollow
                    u, err := getUserInfo(f)
                    if err != nil {
                        log.Printf("Error getting user info [%d]: %s", f, err)
                    } else {
                        msg := fmt.Sprintf("👎 @%s stopped following you.", u.ScreenName())
                        sendDM(usrInfo.IdStr(), msg)
                    }
                }
            }

            // Check follows
            for _, f := range actualFollowers.Ids{
                if IndexOf(previousFollowers.Ids, f) == -1{
                    // Not found -> Unfollow
                    u, err := getUserInfo(f)
                    if err != nil {
                        log.Printf("Error getting user info [%d]: %s", f, err)
                    } else {
                        msg := fmt.Sprintf("👍 @%s started following you!", u.ScreenName())
                        sendDM(usrInfo.IdStr(), msg)
                    }
                }
            }

            // update followers
            previousFollowers = actualFollowers
            log.Printf("Previous followers updated!")
        }
    }
}

func main(){

    // Command line options
    configFile := flag.String("c", "./config.json", "Config file")
    flag.Parse()

    var err error
    config, err = LoadConfig(*configFile)
    if err != nil{
        fmt.Printf("Error loading config [%s]\n", err)
        return
    }

    oauthConfig := &oauth1a.ClientConfig{
        ConsumerKey:    config.ConsumerKey,
        ConsumerSecret: config.ConsumerSecret,
    }
    user := oauth1a.NewAuthorizedConfig(config.AccessToken, config.AccessSecret)

    // Twitter client
    client = twittergo.NewClient(oauthConfig, user)

    mainLoop(config.RefreshTime, config.Username)

}
