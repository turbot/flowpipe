pipeline "test" {

    step "query" {
        sql = "select instance_id, region, _ctx ->> 'connection_name' as connection_name from aws_ec2_instance"
    }

    step "explode" {
        base = steampipe.step.explode
        input = step.query.output.rows
    }

    step "stop_instance" {
        # TODO - call this command or shell?
        base = steampipe.step.command
        # Choose the base container image for this command runner, in this case
        # the AWS CLI.
        image = "amazon/aws-cli"
        # Setting the connection will cause the env vars associated with that
        # plugin type to be automatically set for the container. Env vars are
        # defined per plugin and are documented in the plugin's README. For
        # example, the AWS plugin defines the following env vars:
        # AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_SESSION_TOKEN and
        # AWS_DEFAULT_REGION.
        connection = step.explode.output.connection_name
        # TODO - list of commands?
        # TODO - shell script?
        # TODO - array of command args (common in code)?
        # TODO - Rather than command, just treat this like a function with a shell script in it?
        command = "aws ec2 stop-instance --instance-id '${step.explode.output.instance_id}' --region '${step.explode.output.region}'"
    }

    step "run_function_per_row" {
        base = steampipe.step.function
        # Choose the base container image for this command runner, in this case
        # the AWS CLI.
        location = "./lambda-python"
        # Setting the connection will cause the env vars associated with that
        # plugin type to be automatically set for the container. Env vars are
        # defined per plugin and are documented in the plugin's README. For
        # example, the AWS plugin defines the following env vars:
        # AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_SESSION_TOKEN and
        # AWS_DEFAULT_REGION.
        connection = step.explode.output.connection_name
        # Pass everything to the Lambda function
        input = step.explode.output
        # Set a custom handler (if they are combined)
        handler = "ec2_instance_handler_in_python"
    }

    step "implode" {
        base = steampipe.step.implode
        # Technically this is optional since it is the next step in the file
        # anyway, but for clarity...
        source = step.run_function_per_row
    }

}