const moment = require('moment');

exports.handler = async (event, context) => {
  const currentTime = moment().format('YYYY-MM-DD HH:mm:ss');

  // Printing initial log
  console.log(`Initial log at ${currentTime}`);

  for (let i = 0; i < 2; i++) {
    console.log(`Heartbeat log at ${moment().format('YYYY-MM-DD HH:mm:ss')}`);
    await new Promise(resolve => setTimeout(resolve, 2000));  // wait for 2 seconds
  }

  const response = {
    statusCode: 300,
    body: {
      message: `Nested Hello, World! The current time is ${currentTime}.`,
      event,
      env: process.env,
      context
    },
  };
  return response;
};