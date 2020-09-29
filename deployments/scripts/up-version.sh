#!/bin/bash

# The purpose of this script is to simplify the way of generating a tag for github.
#  Example:
#    * A correction was made and updated the develop branch:
#      * ./deployments/scripts/up-version.sh alpha
#    * A new feature was made and updated the develop branch:
#      * ./deployments/scripts/up-version.sh alpha
#    * We are preparing the branch to develop and send it to production:
#      * ./deployments/scripts/up-version.sh rc
#    * After opening the PR and performing the merge of develop on the master, we must update the master and develop tag:
#      * ./deployments/scripts/up-version.sh minor
#    * We made a correction(hotfix) to the production environment and we have already merged the master branch and automatically need to update the develop branch:
#      * ./deployments/scripts/up-version.sh release
#    * We had to do a refactoring in the services and hear a "breaking changes" in the master branch:
#      * ./deployments/scripts/up-version.sh major

UPDATE_TYPE=$1
BRANCH_NAME=$(git rev-parse --abbrev-ref HEAD)
TAG_VERSION=""

validateUpdateType () {
    case "$UPDATE_TYPE" in
        "alpha") # Used to update an bugfix or an new feature in develop branch
            echo "Update type selected is alpha" ;;
        "rc") # Used when you finish development and start testing in the test environment and in develop branch
            echo "Update type selected is rc(release-candidate)" ;;
        "release") # Used when an correction was applied in master branch
            echo "Update type selected is release" ;;
        "minor") # Used when an new feature is enable in production environment and in master branch
            echo "Update type selected is minor" ;;
        "major") # Used when an big refactor is necessary to breaking changes in master branch
            echo "Update type selected is major" ;;
        *)
            echo "Param Update type is invalid, please use the examples bellow allowed and try again!"
            echo "Params Update type allowed: alpha, rc, release, minor, major"
            exit 1;;
    esac
}

installSemver () {
    chmod +x ./deployments/scripts/install-semver.sh
    ./deployments/scripts/install-semver.sh
}

validateCurrentBranch () {
    # If branch not allowed return errors
    if [[ "$UPDATE_TYPE" == "alpha" || "$UPDATE_TYPE" == "rc" ]]
    then
        if [ "$BRANCH_NAME" != "develop" ]
        then
            echo "Your current branch is \"$BRANCH_NAME\". For this update type only branch enable is \"develop\""
            echo "Please use the follow command to update your project"
            echo "git checkout develop && git pull origin develop"
            exit 1
        fi
        git pull origin develop
    else
        if [ "$BRANCH_NAME" != "master" ]
        then
            echo "Your current branch is \"$BRANCH_NAME\". For this update type only branch enable is \"master\""
            echo "Please use the follow command to update your project"
            echo "git checkout master && git pull origin master"
            exit 1
        fi
        git pull origin master
    fi
}

resetAlphaRcToMaster () {
    alpha_version=$(semver get alpha)
    rc_version=$(semver get rc)
    if [[ "${alpha_version: -2}" != ".0" || "${rc_version: -2}" != ".0" ]]
    then
        echo "Alpha or Release-Candidate found, reseting version to:"
        semver up release
    fi
}

upNewVersion () {
    # Update version
    if [ "$BRANCH_NAME" == "master" ]
    then
        resetAlphaRcToMaster
    fi
    echo "Update version to:"
    semver up "$UPDATE_TYPE"
    if [ "$BRANCH_NAME" == "master" ]
    then
        TAG_VERSION=$(semver get release)
    else
        TAG_VERSION=$(semver get "$UPDATE_TYPE")
    fi
    echo "Tag has been generated in version: $TAG_VERSION"

    # Commit new version
    git tag "$TAG_VERSION"
    git add ".semver.yaml"
    git commit -m "[skip ci] update versioning file"
}

pushChangesAndCheckResponse () {
    # Update version
    git push origin "$BRANCH_NAME"
    RESULT_PUSH=$?
    if [[ $RESULT_PUSH -eq 0 ]]
    then
        git push --tags
        RESULT_PUSH=$?
        if [[ $RESULT_PUSH -eq 0 ]]
        then
            echo "New version was updated!"
        else
            echo "Error on push tags: $RESULT_PUSH"
            exit 1
        fi
    else
        echo "Error on push in branch: $RESULT_PUSH"
        exit 1
    fi
}

validateUpdateType

installSemver

validateCurrentBranch

upNewVersion

pushChangesAndCheckResponse

exit 0