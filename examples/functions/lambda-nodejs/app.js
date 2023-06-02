const moment = require('moment');

exports.handler = async (event, context) => {
  const currentTime = moment().format('YYYY-MM-DD HH:mm:ss');
  const response = {
    statusCode: 200,
    body: {
      message: `Hi, World! The current time is ${currentTime}`,
      event,
      context
    },
  };
  return response;
};