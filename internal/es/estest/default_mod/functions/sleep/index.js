const moment = require("moment");

exports.handler = async (event, context) => {
  const currentTime = moment().format("YYYY-MM-DD HH:mm:ss");

  // Printing initial log
  console.log(`Initial log at ${currentTime}`);

  // Loop for 10 minutes (600 seconds) with a 2-second interval
  for (let i = 0; i < 300; i++) { // 300 iterations * 2 seconds = 600 seconds = 10 minutes
    console.log(`Heartbeat log at ${moment().format("YYYY-MM-DD HH:mm:ss")}`);
    await new Promise((resolve) => setTimeout(resolve, 2000)); // wait for 2 seconds
  }

  const response = {
    statusCode: 300,
    body: {
      message: `Hola, World! The current time is ${currentTime}. From ${event.user.name} with age: ${event.user.age}. Not nested: ${event.notNested}.`,
      event,
      env: process.env,
      context,
    },
  };
  return response;
};
