


NG_URL=$(curl -s localhost:4040/api/tunnels | jq -r '.tunnels[0].public_url')
FP_URL=$(curl -s http://localhost:7103/api/v0/trigger | jq -r '.items[] | select(.name == "integrated_week2_input_demo.trigger.http.priv_cmd").url')
echo
echo $NG_URL/api/v0$FP_URL





```bash
FLOWPIPE_LOG_LEVEL=INFO go run . service start --mod-location ./internal/es/test/demo_mod_input --functions  --log-dir ./tmp --output-dir ./tmp
```

/api/v0/input/slack/foo.bar.z/2342342343


https://console.aws.amazon.com/console/home

pikachu-aaa


