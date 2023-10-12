/*
SPDX-FileCopyrightText: 2020 Amazon.com, Inc. or its affiliates. All Rights Reserved.
SPDX-License-Identifier: MIT-0
*/

process.env = {
  "restrictedActions": "s3:DeleteBucket,s3:DeleteObject"
}

let restrictedActions = process.env.restrictedActions.split(",");
let message = ""
let action = ""

exports.handler = async (event, context) => {
    var policyObject = event.policy
    let policyActions = policyObject.Statement[0].Action

    //const found = policyActions.some(restrictedActions)

    var found=0;
    for(i=0; i < restrictedActions.length; i++  ){
      if(policyActions.indexOf(restrictedActions[i]) >= 0){
         found =1;
         break
      }
    }
    //const found = policyActions.some(r=> policyActions.indexOf(restrictedActions) >= 0)

    if (found) {
      message = `Policy was changed with restricted actions: ${restrictedActions}`
      action = "remedy"
    } else {
      message = `Policy was changed to: ${event.policy}`
      action = "alert"
    }

    return {
      "message": message,
      "action": action
    }
};
