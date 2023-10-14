exports.handler = async (event, context) => {
    var transformedData = {
      "policy": JSON.parse(event.detail.requestParameters.policyDocument),
      "policyMeta": {
        "arn": event.detail.responseElements.policy.arn,
        "policyName": event.detail.responseElements.policy.policyName,
        "defaultVersionId": event.detail.responseElements.policy.defaultVersionId
      },
    }

    console.log("Event: ", JSON.stringify(transformedData))

    return transformedData
};
