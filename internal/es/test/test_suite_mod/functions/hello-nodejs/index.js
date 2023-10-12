const moment = require('moment');

exports.handler = async (event, context) => {
  console.log('event', event)
  console.log('context', context)
  const currentTime = moment().format('YYYY-MM-DD HH:mm:ss');
  const response = {
    statusCode: 200,
    body: {
      message: `Hola, World! The current time is ${currentTime}. From ${event.user.name} with age: ${event.user.age}. Not nested: ${event.notNested}.`,
      event,
      context
    },
  };
  return response;
};