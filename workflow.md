My workflow looks like this (kept here for my personal convenience):

    # make a fresh sd card with raspbian jessie lite
    ./flash 2017-04-10-raspbian-jessie-lite.zip

    # build and copy appliance, changing the default hostname to 'pi1'
    ./build-and-copy pi1
    
    #... boot the pi

    # login via ssh
    ssh pi@pi1

    # install docker
    curl -sSL https://get.docker.com/ | sh
    curl -fsSL https://test.docker.com/ | sh
