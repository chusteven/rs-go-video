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
100   158  100   158    0     0  69911      0 --:--:-- --:--:-- --:--:-- 79000
{
  "videos": [
    "/Users/stevenchu/Downloads/tongue-singing-choir.mp4"
  ],
  "displays": [
    {
      "screen_id": 1,
      "width": 1440,
      "height": 900,
      "origin_x": 0,
      "origin_y": 0,
      "index": 0
    }
  ]
}
$
```

Then submit a POST request to actually play one:
```bash
$ curl -XPOST http://localhost:8080/video -d'{"video": "/Users/stevenchu/Downloads/tongue-singing-choir.mp4", "screen_id": 1}' | jq .
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   191  100   111  100    80  42857  30888 --:--:-- --:--:-- --:--:-- 95500
{
  "status": "success",
  "message": "playing video /Users/stevenchu/Downloads/tongue-singing-choir.mp4 on screen 1"
}
$
```