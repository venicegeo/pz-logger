#!/usr/bin/env bash
sudo apt-get update
sudo apt-get upgrade

# install openjdk-7 
sudo apt-get purge openjdk*
sudo apt-get -y install openjdk-7-jdk

#alias list="ls -a"
 
#get golang
cd /usr/local
sudo wget https://storage.googleapis.com/golang/go1.6.2.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.6.2.linux-amd64.tar.gz

#add go to path
#cd /etc/profile.d
#sudo echo "export PATH=$PATH:/usr/local/go/bin" >> /etc/profile.d/gosetup.sh
echo 'export PATH=$PATH:/usr/local/go/bin' >>/home/vagrant/.bash_profile

export PATH=$PATH:/usr/local/go/bin

echo "INSTALLED THE GO 1111111111111111111111111111111111111111111111111 "
echo $PATH

#sudo apt-get install git-all
sudo apt-get install git htop -y -q

#build and start the go app
cd /vagrang/logger


#go get github.com/venicegeo/pz-logger/logger
#go install github.com/venicegeo/pz-logger/logger