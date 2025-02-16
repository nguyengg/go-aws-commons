#!/usr/bin/env python3

import argparse
import boto3
import os
import subprocess
import sys
import zipfile


# TODO figure out how build.py can auto-update itself.
# To get or update to latest content run:
# curl --proto '=https' -fo build.py https://raw.githubusercontent.com/nguyengg/go-aws-commons/main/lambda/build.py

# boto3 should already be installed if building with these images.
# https://github.com/aws/aws-codebuild-docker-images/blob/master/al2/aarch64/standard/3.0/Dockerfile
# https://github.com/aws/aws-codebuild-docker-images/blob/master/al2/x86_64/standard/5.0/Dockerfile
# https://github.com/aws/aws-codebuild-docker-images/blob/master/ubuntu/standard/7.0/Dockerfile
def main():
    parser = argparse.ArgumentParser(
        prog='build.py',
        description='Build a Lambda Go handler and update the associated function with the compressed build artifacts.',
        epilog="""The script can be updated with `curl --proto '=https' -fo build.py https://raw.githubusercontent.com/nguyengg/go-aws-commons/main/lambda/build.py`""")
    parser.add_argument('--assume-role', dest='role_arn', type=str, metavar='arn:aws:iam::123456789012:role/my-role',
                        help='If --assume-role is given a role ARN, this role will be assumed to produce the '
                             'credentials that are used to update Lambda functions.')
    parser.add_argument('-b', '--build', action='store_true',
                        help='If -b is given, `go build` is always executed. If -b is present but not -u, stop after '
                             'build step; the build artifacts are available in --bin-dir (default to ./bin). If neither '
                             '-b nor -u are given then both actions take place in sequence (implicit -bu).')
    parser.add_argument('-u', '--update', action='store_true',
                        help='If -u is given, the function whose name is provided by --function-name or main_package '
                             'will be updated using the build artifacts produced by -b. If build output is not '
                             'available, -b is implicitly added. If neither -b nor -u are given then both actions take '
                             'place in sequence (implicit -bu).')
    parser.add_argument('-f', '--function', dest='functions', action='append', default=[], metavar='function-name',
                        help='-f/--function may be given multiple times to provide the names, ARNs, or partial ARNs of '
                             'the functions to be updated. If not given, the basename of required positional argument '
                             '(main_package) will be used as the function name.')
    parser.add_argument('-d', '--delete', action='store_true',
                        help='If -d is given, the executables produced by `go build` will be deleted only if they were '
                             'produced by the command, and if the command completes successfully.')
    parser.add_argument('--load-dotenv', action=LoadDotEnvAction, metavar='/path/to/.env',
                        help='If --load-dotenv is given with an optional .env path, load_env from dotenv (must be '
                             'preinstalled) will be used to set additional environment variables that apply to build '
                             'and update steps. Shell variables and those provided with --env-var will not be '
                             'overridden due to default behaviour of load_dotenv(override=False).')
    parser.add_argument('-e', '--env-var', action=EnvVarAction, metavar='AWS_PROFILE=my-profile',
                        help='-e/--env-var may be specified multiple times to provide additional environment variables '
                             'that apply to build and update steps. These environment variables will override shell '
                             'variables as well as those loaded with --load-dotenv.')
    parser.add_argument('--tags', default='lambda.norpc', metavar='tag,list', nargs='?', const=None,
                        help='Override the comma-separated list of build tags passed to `go build`. By default, '
                             '`-tags lambda.norpc` is provided. To pass no tags, specify `--tags` without any value.')
    parser.add_argument('--bin-dir', default='./bin/', metavar='./bin/',
                        help='Change the output directory, default to ./bin/ .')
    parser.add_argument('main_package',
                        help='The directory that contains an executable Go package (package name should be main, '
                             'and one of the files in the directory must have a main() method). If passed an archive '
                             'with .zip extension, -u is implied while -b must not be given, and the ZIP file will be '
                             'used to update function code.')

    args = parser.parse_args()

    package_name = os.path.basename(args.main_package)
    build = args.build
    update = args.update
    functions = list(args.functions)
    if len(functions) == 0:
        functions.append(os.path.splitext(package_name)[0])

    if package_name.endswith('.zip'):
        if build:
            print("cannot specify -b if zip file is given")
            sys.exit(1)
        update_and_wait(functions, args.main_package)
        return

    if not build and not update:
        build = True
        update = True

    output = os.path.join(args.bin_dir, package_name)
    if not build and not os.path.exists(output):
        build = True

    if build:
        cmd = ["go", "build"] + (["-tags", args.tags] if args.tags else []) + ["-o", output, args.main_package]
        print(f"building {args.main_package} to {output}:", ' '.join(cmd))
        p = subprocess.run(
            cmd,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            text=True)
        if p.returncode != 0:
            print(f"build failed with exit code {p.returncode}")
            if p.stdout:
                print(p.stdout)
            sys.exit(1)

    if update:
        archive = os.path.join(args.bin_dir, package_name + ".zip")
        with zipfile.ZipFile(archive, 'w', zipfile.ZIP_DEFLATED) as f:
            f.write(output, "bootstrap")

        update_and_wait(functions, archive, args.role_arn)

        # only delete if we've done any function update.
        if args.delete and build:
            print(f"deleting {output}")
            os.remove(output)


def update_and_wait(functions, file, role_arn=None):
    if role_arn:
        sts_client = boto3.client('sts')
        response = sts_client.assume_role(RoleArn=role_arn, RoleSessionName="EnforceLogGroupsRetention")
        credentials = response['Credentials']
        client = boto3.client('lambda',
                              aws_access_key_id=credentials['AccessKeyId'],
                              aws_secret_access_key=credentials['SecretAccessKey'],
                              aws_session_token=credentials['SessionToken'])
    else:
        client = boto3.client('lambda')

    for function_name in functions:
        print(f"updating function {function_name} with {file}")
        with open(file, 'rb') as f:
            client.update_function_code(
                FunctionName=function_name,
                ZipFile=f.read(),
            )

    for function_name in functions:
        print(f"waiting for function {function_name} to be updated")
        client.get_waiter('function_updated_v2').wait(FunctionName=function_name)


class EnvVarAction(argparse.Action):
    def __init__(self, option_strings, dest, nargs=None, **kwargs):
        if nargs is not None:
            raise ValueError(f"nargs {nargs} must be None")
        super().__init__(option_strings, dest, nargs=None, **kwargs)

    def __call__(self, parser, namespace, values, option_string=None):
        key, _, value = values.partition("=")
        if not value:
            raise ValueError(f"not in format KEY=value: {values}")
        os.environ[key] = value


class LoadDotEnvAction(argparse.Action):
    def __init__(self, option_strings, dest, nargs=None, **kwargs):
        if nargs is not None and nargs != '?':
            raise ValueError(f"nargs {nargs} must be ? if given")
        super().__init__(option_strings, dest, nargs='?', **kwargs)

    def __call__(self, parser, namespace, value, option_string=None):
        from dotenv import load_dotenv
        load_dotenv(dotenv_path=value)


if __name__ == "__main__":
    main()
