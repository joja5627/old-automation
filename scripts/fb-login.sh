#!/bin/bash

# If it redirects to http://www.facebook.com/login.php at the end, wait a few minutes and try again

EMAIL='joja5627@gmail.com' # edit this
PASS='Cu112145@buff' # edit this

COOKIES='cookies.txt'
USER_AGENT='Firefox/3.5'



curl -X GET 'https://www.facebook.com/home.php' --verbose --user-agent $USER_AGENT -j -b cookies.txt --location https://login.facebook.com/login.php
# 
#--cookie-jar $COOKIES
curl -X POST 'https://login.facebook.com/login.php' --verbose --user-agent 'Firefox/3.5' --data-urlencode "email=joja5627@gmail.com" --data-urlencode "pass=Cu112145@buff" --cookie "cookies.txt" --cookie-jar "cookies.txt" 
curl -X GET 'https://www.facebook.com/home.php' --verbose --user-agent $USER_AGENT --cookie $COOKIES --cookie-jar $COOKIES -o request_3.html

