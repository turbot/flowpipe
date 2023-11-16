import requests

def my_handler(event, context):
    response = requests.get("https://api.spacexdata.com/v4/launches/upcoming?limit=5")
    launches = response.json()

    launch_data = []
    for launch in launches:
        launch_data.append({
            "mission_name": launch["name"],
            "launch_date_utc": launch["date_utc"],
            "details": launch["details"]
        })

    return {
        "statusCode": 200,
        "body": launch_data
    }
