# dpdk-go

## Golang高性能轻量级网络包收发框架

### dpdk环境搭建

```shell
# 建议使用Ubuntu18.04或Ubuntu20.04

# 安装numactl
unzip numactl-master.zip
cd numactl-master
./autogen.sh
./configure
make -j 8
make install
ldconfig

# 安装dpdk-18.11.11
tar -xvf dpdk-18.11.11.tar.xz
cd dpdk-stable-18.11.11
# 环境变量
export RTE_SDK="/root/dpdk-stable-18.11.11/"
export RTE_TARGET="x86_64-native-linuxapp-gcc"
# 编译DPDK
make config T=$RTE_TARGET
make -j 8 install T=$RTE_TARGET

# 配置dpdk

# UIO和VFIO二选一

# UIO
modprobe uio
insmod /root/dpdk-stable-18.11.11/x86_64-native-linuxapp-gcc/kmod/igb_uio.ko
/root/dpdk-stable-18.11.11/usertools/dpdk-devbind.py --bind=igb_uio eth0

# VFIO Aliyun适用
# 参考链接
https://help.aliyun.com/document_detail/310880.htm
cat /proc/cmdline
vim /etc/default/grub
# "GRUB_CMDLINE_LINUX" 追加 "intel_iommu=on"
update-grub2
reboot
# 检查是否开启成功
cat /proc/cmdline
modprobe vfio && modprobe vfio-pci
echo 1 > /sys/module/vfio/parameters/enable_unsafe_noiommu_mode
ethtool -i eth0
# vfio-pci后的参数即为ethtool所查看的网卡的pcie设备号
/root/dpdk-stable-18.11.11/usertools/dpdk-devbind.py -b vfio-pci 0000:00:05.0

# KNI
insmod /root/dpdk-stable-18.11.11/x86_64-native-linuxapp-gcc/kmod/rte_kni.ko carrier=on

# 内存大页
echo 1024 > /sys/devices/system/node/node0/hugepages/hugepages-2048kB/nr_hugepages
mkdir -p /mnt/huge_2M
mount -t hugetlbfs none /mnt/huge_2M -o pagesize=2M
echo 0 > /sys/devices/system/node/node0/hugepages/hugepages-1048576kB/nr_hugepages
mkdir -p /mnt/huge_1G
mount -t hugetlbfs none /mnt/huge_1G -o pagesize=1G
```

### dpdk-go编译

```shell
cd /root
git clone https://github.com/FlourishingWorld/dpdk-go.git
cd dpdk-go
mkdir bin
# 确保CGO开启并添加以下环境变量
export CGO_CFLAGS_ALLOW=.*
export CGO_LDFLAGS_ALLOW=.*
go build -o /root/dpdk-go/bin/dpdk_go dpdk-go/cmd
```

### 如何使用

```go
// 通过RX和TX两个管道可发送和接收原始以太网报文
pkt := <-dpdk.DPDK_RX_CHAN
dpdk.DPDK_TX_CHAN <- pkt
```
