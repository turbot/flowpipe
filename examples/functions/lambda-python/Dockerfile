FROM public.ecr.aws/lambda/python:3.10

# Set the working directory inside the container
WORKDIR ${LAMBDA_TASK_ROOT}

# Install the function's dependencies using file requirements.txt
# from your project folder.
COPY requirements.txt  .
RUN  pip3 install -r requirements.txt --target .

# Copy the rest of the application code
COPY . .

# Set the CMD to your handler
CMD [ "app.handler" ]