export OS_AUTH_URL="http://keystone.local/v3"
export OS_IDENTITY_API_VERSION=3
export OS_PROJECT_NAME="my-project"
export OS_PROJECT_DOMAIN_NAME="domain"
export OS_USERNAME="my-username"
export OS_USER_DOMAIN_NAME="domain"
echo "Please enter your OpenStack Password: "
read -sr OS_PASSWORD_INPUT
export OS_PASSWORD="${OS_PASSWORD_INPUT}"
export OS_REGION_NAME="region"
