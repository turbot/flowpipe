


NG_URL=$(curl -s localhost:4040/api/tunnels | jq -r '.tunnels[0].public_url')
FP_URL=$(curl -s http://localhost:7103/api/v0/trigger | jq -r '.items[] | select(.name == "integrated_week2_input_demo.trigger.http.priv_cmd").url')
echo
echo $NG_URL/api/v0$FP_URL




FLOWPIPE_LOG_LEVEL=INFO go run . service start --mod-location ~/src/int2023_week2/demo_mod/ --functions