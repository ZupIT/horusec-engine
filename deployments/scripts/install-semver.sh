#!/bin/bash

checkIfInstallationIsValid () {
    semver &> /dev/null
    RESPONSE=$?
    if [ $RESPONSE != "0" ]
    then
        LOCATION_SEMVER=$(which semver)
        echo "Semver is not installed please remove the binary in location [$LOCATION_SEMVER] and run again"
        exit 1
    else
        echo "Semver was installed with success!"
    fi
}

installSemver () {
    semver &> /dev/null
    RESPONSE=$?
    if [ $RESPONSE != "0" ]
    then
        echo "Installing semver..."
        curl https://horus-assets.s3.amazonaws.com/semver -o ./semver
        chmod +x ./semver
        sudo mv ./semver /usr/local/bin/semver
    fi
}

installSemver

checkIfInstallationIsValid