const moment = require('moment');

exports.handler = async (event, context) => {
  console.log("")
  console.log("")
  console.log("")
  console.log('event', event)
  console.log('context', context)
  console.log("")
  console.log("")
  console.log("")
  const currentTime = moment().format('YYYY-MM-DD HH:mm:ss');
  const response = {
    statusCode: 200,
    body: {
      message: `Hola, World! The current time is ${currentTime}. From ${event.user.name} with age: ${event.user.age}. Not nested: ${event.notNested}.`,
      event,
      env: process.env,
      context
    },
  };
  return response;
};