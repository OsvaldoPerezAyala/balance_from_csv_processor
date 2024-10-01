# Project balance processor from CSV
The proposal of this project is, by means of a mounted file into the project, 
summarize the transactions into the csv file, generating a mail with the transaction summary, 
counting transactions by month and attach the file. Additional saving transactions into mongo database 
in order to will be exploded in the future. 

## 📝 Execution requirements
- Golang 1.20 +
- Docker and Docker desktop 
>**Note**: This project is prepared to work with docker compose  

## 🔨 To start this application  
Download the repo and localize the `docker-compose.yml`, open in your plain text favorite editor.<br>
Replace the value of the variables `USER_EMAIL` and `APPLICATION_KEY` with your own.<br>
It is designed to work with Gmail, and you will need to enable an application key.<br>

Once replaced, enter the balance_from_csv_processor folder and run the following commands in your terminal:
`go mod tidy` and then `docker compose up --build`
>**Note**: Make sure you have permission to run the second command or use administrator privileges.

## ⚙️ 🔍 To perform tests
Locally with Docker: 

Below I add the CURL of the service to use it from postman.

### /summary/csv

> curl --location 'localhost:8080/summary/csv' \
--header 'Content-Type: application/json' \
--data-raw '{"email":"your email address"}'

## ⚙️ 🔍☁️  Tests in the cloud

Additionally, the service was deployed on AWS and can be tested with the following curl command:

## /develop/summary

> curl --location 'https://8t4xic7m4h.execute-api.us-east-1.amazonaws.com/develop/summary' \
--header 'Content-Type: application/json' \
--data-raw '{"email":"your email address"}'

>**Note**: 📨 In the field email into the payload, replace the value “your email address” for your recipient’s email address.
Additionally, for cloud testing, the email will be sent from the address osvaldofpa@gmail.com previously configured.






