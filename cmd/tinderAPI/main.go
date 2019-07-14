package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/antonholmquist/jason"
	"github.com/golang/oauth2"
)

//   "https://www.facebook.com/v2.6/dialog/oauth?redirect_uri=fb464891386855067%3A%2F%2Fauthorize%2F&display=touch&state=%7B%22challenge%22%3A%22IUUkEUqIGud332lfu%252BMJhxL4Wlc%253D%22%2C%220_auth_logger_id%22%3A%2230F06532-A1B9-4B10-BB28-B29956C71AB1%22%2C%22com.facebook.sdk_client_state%22%3Atrue%2C%223_method%22%3A%22sfvc_auth%22%7D&scope=user_birthday%2Cuser_photos%2Cuser_education_history%2Cemail%2Cuser_relationship_details%2Cuser_friends%2Cuser_work_history%2Cuser_likes&response_type=token%2Csigned_request&default_audience=friends&return_scopes=true&auth_type=rerequest&client_id=464891386855067&ret=login&sdk=ios&logger_id=30F06532-A1B9-4B10-BB28-B29956C71AB1&ext=1470840777&hash=AeZqkIcf-NEW6vBd"

// "https://www.facebook.com/v3.3/dialog/oauth?
//     client_id={app-id}
//     &display=popup
//     &response_type=token
//     &redirect_uri=ms-app://{package-security-identifier}"
const (
	mobileUserAgent = "Tinder/7.5.3 (iPhone; iOS 10.3.2; Scale/2.00)"
	authURL         = "https://www.facebook.com/v3.3/dialog/oauth?client_id={app-id}&display=popup&response_type=token&redirect_uri=ms-app://{package-security-identifier}"
	username        = "3035015076"
	password        = "Cu112145@buff"
	host            = "https://api.gotinder.com"
)

var (

//accestoken = fb_auth_token.get_fb_access_token(fb_username, fb_password)
//id = fb_auth_token.get_fb_id(fb_access_token)

)

// def get_fb_access_token(email, password):
//     s = robobrowser.RoboBrowser(user_agent=MOBILE_USER_AGENT, parser="lxml")
//     s.open(FB_AUTH)
//     f = s.get_form()
//     f["pass"] = password
//     f["email"] = email
//     s.submit_form(f)
//     f = s.get_form()
//     try:
//         s.submit_form(f, submit=f.submit_fields['__CONFIRM__'])
//         access_token = re.search(
//             r"access_token=([\w\d]+)", s.response.content.decode()).groups()[0]
//         return access_token
//     except Exception as ex:
//         print("access token could not be retrieved. Check your username and password.")
//         print("Official error: %s" % ex)
//         return {"error": "access token could not be retrieved. Check your username and password."}

// def get_fb_id(access_token):
//     if "error" in access_token:
//         return {"error": "access token could not be retrieved"}
//     """Gets facebook ID from access token"""
//     req = requests.get(
//         'https://graph.facebook.com/me?access_token=' + access_token)
//     return req.json()["id"]

// headers = {
//     'app_version': '6.9.4',
//     'platform': 'ios',
//     "content-type": "application/json",
//     "User-agent": "Tinder/7.5.3 (iPhone; iOS 10.3.2; Scale/2.00)",
// 	"X-Auth-Token": config.tinder_token,
// }

type AccessToken struct {
	Token  string
	Expiry int64
}

func readHttpBody(response *http.Response) string {

	fmt.Println("Reading body")

	bodyBuffer := make([]byte, 5000)
	var str string

	count, err := response.Body.Read(bodyBuffer)

	for ; count > 0; count, err = response.Body.Read(bodyBuffer) {

		if err != nil {

		}

		str += string(bodyBuffer[:count])
	}

	return str

}

// "https://www.facebook.com/v3.3/dialog/oauth?
//     client_id={app-id}
//     &display=popup
//     &response_type=token
//     &redirect_uri=ms-app://{package-security-identifier}"

//Converts a code to an Auth_Token
func GetAccessToken(client_id string, code string, secret string, callbackUri string) AccessToken {
	fmt.Println("GetAccessToken")
	//https://graph.facebook.com/oauth/access_token?client_id=YOUR_APP_ID&redirect_uri=YOUR_REDIRECT_URI&client_secret=YOUR_APP_SECRET&code=CODE_GENERATED_BY_FACEBOOK
	response, err := http.Get("https://graph.facebook.com/oauth/access_token?client_id=" +
		client_id + "&redirect_uri=" + callbackUri +
		"&client_secret=" + secret + "&code=" + code)

	if err == nil {

		auth := readHttpBody(response)

		var token AccessToken

		tokenArr := strings.Split(auth, "&")

		token.Token = strings.Split(tokenArr[0], "=")[1]
		expireInt, err := strconv.Atoi(strings.Split(tokenArr[1], "=")[1])

		if err == nil {
			token.Expiry = int64(expireInt)
		}

		return token
	}

	var token AccessToken

	return token
}

func Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// generate loginURL
	fbConfig := &oauth2.Config{
		// ClientId: FBAppID(string), ClientSecret : FBSecret(string)
		// Example - ClientId: "1234567890", ClientSecret: "red2drdff6e2321e51aedcc94e19c76ee"

		ClientID:     "", // change this to yours
		ClientSecret: "",
		RedirectURL:  "http://<domain name and don't forget port number if you use one>/FBLogin", // change this to your webserver adddress
		Scopes:       []string{"email", "user_birthday", "user_location", "user_about_me"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.facebook.com/dialog/oauth",
			TokenURL: "https://graph.facebook.com/oauth/access_token",
		},
	}

	url := fbConfig.AuthCodeURL("")

	// Home page will display a button for login to Facebook

	w.Write([]byte("<html><title>Golang Login Facebook Example</title> <body> <a href='" + url + "'><button>Login with Facebook!</button> </a> </body></html>"))
}

func FBLogin(w http.ResponseWriter, r *http.Request) {
	// grab the code fragment

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	code := r.FormValue("code")

	ClientId := "" // change this to yours
	ClientSecret := ""
	RedirectURL := "http://<domain name and don't forget port number if you use one>/FBLogin"

	accessToken := GetAccessToken(ClientId, code, ClientSecret, RedirectURL)

	response, err := http.Get("https://graph.facebook.com/me?access_token=" + accessToken.Token)

	// handle err. You need to change this into something more robust
	// such as redirect back to home page with error message
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	str := readHttpBody(response)

	// dump out all the data
	// w.Write([]byte(str))

	// see https://www.socketloop.com/tutorials/golang-process-json-data-with-jason-package
	user, _ := jason.NewObjectFromBytes([]byte(str))

	id, _ := user.GetString("id")
	email, _ := user.GetString("email")
	bday, _ := user.GetString("birthday")
	fbusername, _ := user.GetString("username")

	w.Write([]byte(fmt.Sprintf("Username %s ID is %s and birthday is %s and email is %s<br>", fbusername, id, bday, email)))

	img := "https://graph.facebook.com/" + id + "/picture?width=180&height=180"

	w.Write([]byte("Photo is located at " + img + "<br>"))
	// see https://www.socketloop.com/tutorials/golang-download-file-example on how to save FB file to disk

	w.Write([]byte("<img src='" + img + "'>"))
}

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", Home)
	mux.HandleFunc("/FBLogin", FBLogin)

	http.ListenAndServe(":8080", mux)
}
