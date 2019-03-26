#!/usr/bin/env bash
set -e

# Copies any default config from this bootstrapper image to the configuration empty-dir (overwriting any defaults from
# the user-specified cassandra image)
cp -rLv ${CONF_STAGING_DIRECTORY}/* /configuration

# Copies any custom configuration (if any has been specified) from the user's config-map to the same empty-dir
# (overwriting the above)
if [[ -d /custom-config && ! -z `ls -A /custom-config` ]] ; then
    cp -rLv /custom-config/* /configuration
fi

# Copies any extra libraries from this bootstrapper image to the extra-lib empty-dir
cp -v ${LIB_STAGING_DIRECTORY}/* /extra-lib

# Run the bootstrapper application, which will rewrite config as needed within /configuration
java -Xmx64m -jar /cassandra-bootstrapper.jar
