#!/bin/bash

checkIfInstallationIsValid () {
    EXPECTED_RESPONSE="Semantic version tool helper to validate and increase versions semantically"
    EXISTS_SEMVER=$(semver | grep "$EXPECTED_RESPONSE")
    if [[ -z "$EXISTS_SEMVER" ]]
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
        go get -u github.com/wiliansilvazup/semver-cli/cmd/semver
    fi
}

installSemver

checkIfInstallationIsValid
