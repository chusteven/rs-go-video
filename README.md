# Video server

Run the server with
```bash
$ go run main.go
```

Then submit a GET request to see what screens and videos are available:
```bash
$ curl -XGET http://localhost:8080/ | jq .
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   241  100   241    0     0   180k      0 --:--:-- --:--:-- --:--:--  235k
{
  "videos": [
    "./videos/jesus.webm",
    "./videos/tongue-singing.webm"
  ],
  "displays": [
    {
      "screen_id": 1,
      "width": 1440,
      "height": 900,
      "origin_x": 0,
      "origin_y": 0,
      "index": 0
    },
    {
      "screen_id": 2,
      "width": 3440,
      "height": 1440,
      "origin_x": -953,
      "origin_y": 900,
      "index": 1
    }
  ]
}
$
```

Then submit a POST request to actually play one:
```bash
$ $ curl -XPOST http://localhost:8080/video -d'{"video": "./videos/jesus.webm", "screen_id": 2}' | jq .
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   127  100    79  100    48  72878  44280 --:--:-- --:--:-- --:--:--  124k
{
  "status": "success",
  "message": "playing video ./videos/jesus.webm on screen 2"
}
$
```