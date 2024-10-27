# Backend Technical Test at Scalingo

This is the result of my try at the Scalingo Backend test. The goal was to create an API 
responsible for retrieving the 100 most recent repositories created from the GitHub API. Information for each repository
had to be treated concurrently then used for specific query filters. \
To ensure good performance, scalability and ease of use: 
* Uses concurrent processing to fetch and aggregate data from GitHub, improving efficiency and handling high 
request volumes more effectively.
* Implements in-memory caching to store the results of common queries, reducing redundant GitHub API calls and 
accelerating response times for frequent requests.
* Implements direct search queries with multiple parameters, allowing users to retrieve 
repositories that match specific criteria in real time.

## Endpoints
* /ping \
Check the server status with a simple ping.
```
{ "status": "pong" }
```

* /repos \
100 last GitHub repositories. After first getting requestion for the list of repositories, we use goroutines to fetch 
more information for each one.
```json
{
  "name": "leaflet",
  "owner": "laurisasastoque",
  "url": "https://api.github.com/repos/laurisasastoque/leaflet",
  "description": "A leaflet project in html ",
  "license": "mit",
  "language": {
    "go": 7683,
    "javascript": 369423
  }
}
```


## Filtering
User can specify ``language`` and ``license`` parameters when requesting for the repositories:  \
``/repos?language=css``
```json
[{
  "name": "Governates",
  "owner": "Ibrahim20065",
  "url": "https://api.github.com/repos/Ibrahim20065/Governates",
  "language": {
    "css": 583,
    "html": 4322
  }
},
  {
    "name": "ph_1b10_rafi_assignment_07",
    "owner": "hasanRafi2002",
    "url": "https://api.github.com/repos/hasanRafi2002/ph_1b10_rafi_assignment_07",
    "language": {
      "css": 58,
      "html": 472,
      "javascript": 23522
    }
  },
  ...
```


``/repos?language=html&license=apache-2.0``
```json
[{
  "name": "iehr-client-external-idp-demo",
  "owner": "iehr-ai",
  "url": "https://api.github.com/repos/iehr-ai/iehr-client-external-idp-demo",
  "license": "apache-2.0",
  "language": {
    "html": 309,
    "typescript": 10970
  }
},
...
```



## Execution

### Setting environment variable for authentification
I made it obligatory to have an authenticated client. Create a ``.env`` file and insert a token:
```
GITHUB_TOKEN = "ghp_personnal_github_token_here"
```
### Launch server container
```
docker compose up
```

Application will be then running on port `5000`

