FROM public.ecr.aws/lambda/nodejs:20

# Set the working directory inside the container
WORKDIR ${LAMBDA_TASK_ROOT}

# Copy function code
COPY package*.json ./

# Install dependencies
RUN npm install --omit=dev

# Copy the rest of the application code
COPY . .

# Set the CMD to your handler
CMD [ "app.handler" ]