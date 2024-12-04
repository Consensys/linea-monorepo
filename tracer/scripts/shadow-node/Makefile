ROOT = $(shell pwd)
BESU = linea-besu
PLUGIN = linea-arithmetization
CONSTRAINTS = ${PLUGIN}/zkevm-constraints
CORSET = corset

.ONESHELL:

.PHONY:

update-corset:
        cd ${CORSET}
        git pull
        cargo install --path .

clean-plugin:
        rm -rf ${BESU}/plugins
        mkdir -p ${BESU}/plugins

update-plugin: clean-plugin
        cd ${PLUGIN}
        rm -f arithmetization/build/libs/*jar
        ./gradlew --no-daemon jar && cp arithmetization/build/libs/*.jar ${ROOT}/${BESU}/plugins

update-constraints:
        cd ${PLUGIN}
        git submodule update zkevm-constraints
        $(MAKE) -C zkevm-constraints -B zkevm.bin
        mv zkevm-constraints/zkevm.bin ${ROOT}

update: update-plugin update-corset update-constraints

restart:
        sudo systemctl stop besu
        sudo rm /ramdisk/*lt
        sudo systemctl start besu
        tail -f besu.log

restart-testnet:
        sudo systemctl stop besu-testnet
        sudo rm /ramdisk/testnet-traces/*lt
        sudo systemctl start besu-testnet
        tail -f besu-testnet.log

shadow-node: update restart

status:
        sudo systemctl status besu

log:
        tail -f besu.log

log-testnet:
        tail -f besu-testnet.log
