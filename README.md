## 自己动手写Docker（书籍的精简版和一些小Tip，比如如何在windows下编写等）
### Namespace 作隔离
1. https://learnku.com/articles/42072
2. 原理其实就是创建进程的时候通过配置sysCall.CLONE_NEWUTS等配置来为进程创建Namepace,从而实现隔离
3. 补充下定义（写成文章）

### Cgroup 作资源限制
1. https://learnku.com/articels/42117
2. 原理主要是：系统 默认已经 为每个subsystem创建了一个默认的hierarchy,它在linux的/sys/fs/cgroup路径下，想限制某个进程 ID的内存，就在/sys/
fs/cgroup/memory文件夹下创建一个限制memeory的cgroup，方式就是在空上目录下创建一个目录，kernel自动会把该文件夹标记为一个cgroup；创建好后，修改里面的tasks文件，
   将进程ID写入，然后再修改其memory.limit_in_bytes文件，就能达到限制该进程的内存使用了
3. 如果想删除cgroup创建出来的目录，需要先umount <dir-name>
4. 主要有三个概念
   - cgroup hierarchy中的节点，用于管理进程和subsystem的控制关系
   - subsystem作用于hierarchy上的cgroup节点，并控制节点中进程的资源占用
   - hierarchy将cgroup通过树状结构串起来，并通过虚拟文件系统的方式暴露给用户

### Union File System
1. 简称是UnionFS，一种可以把其他文件系统联合到一个联合挂载点的文件系统服务，当我们对这个联合文件系统进行写操作的时候，系统是真正写到了一个新的文件，
这样看下来这个虚拟的联合文件系统是可以对任何文件进行操作的，但是其实它并没有改变原来的文件，这其实就是“写时复制”，一种资源管理技术；它的思想是
   如果一个资源是重复的，但没有任何修改，这时并不需要立即创建一个新的资源，这个资源 可以被新旧实例共享。创建新资源发生在第一次写操作时。通过这种方式，
   可以显著减少对未修改资源复制所带来的消耗。
   
2. Docker的存储方式有AUFS、overlayFS等。默认的是overlay2

### 构造容器
#### tag-3.1 构造实现run命令版本的容器分支
0. 目标：运行./docker run -ti /bin/sh; ./docker run -ti /bin/ls
1. 对着书把代码里面的注释补充下
2. 构建go build -o mydocker .
3. 出现fork/exec /proc/self/exe: no such file or directory： github.com/xianlubird/mydocker/issues/33; 需要输入 mount -t proc proc /proc
4. 调试：dlv exec ./docker -- run -ti /bin/bash
5. 大致的流程：解析到run的时候会调用runCommand里面的action，这个action会调用根据参数调用initCommand，初始化容器资源，运行输入的命令/bin/sh

### tag-3.2 增加容器资源限制
0. 目标：运行./docker run -ti -m 100m -cpuset 1 -cpushare 512 /bin/sh 的方式来限制容器的内存和CPU配置
1. 先实现cgroup的操作逻辑，在cgroup/subsystem下把相关的配置实现，然后在使用cgroup_manager把它们管理起来建立与容器的关系(cgroup会有多个subsystem)
3. 当我们运行./docker run -ti -m 100m /bin/sh时出现：no space left on device，查看github.com/xianlubird/mydocker/issues/74,解决的方式是在写入pid之前，先检查下相关的配置项是否有配置值

### 构建镜像
#### tag-4.1
1. 目标：前面章节实现的容器的目录还是当前运行程序的目录，本节的目标在于基于busybox给我们的容器提供文件系统
2. pivot_root: 是一个系统调用，主要功能是去改变当前的root文件系统; 
3. 运行./docker run -ti /bin/sh
4. 步骤
    - 准备镜像层目录：`docker pull busybox`, `docker run -d busybox top -b`，`docker export -o busybox.tar <contain-id>`
    `tar -xvf busybox.tar -C /root/docker-exp/busybox`
    - 写完代码后，`go build .`, `./docker run -ti /bin/sh`
5. 问题
    - 直接按书本的代码在pivot_root系统调用那里报错Invalid argument, https://github.com/xianlubird/mydocker/issues/13


#### tag-4.2
1. 目标：在4.1章节中，我们实现了使用鹤机/root/docker-exp/busybox目录作为文件的根目录，但在容器内对文件的操作仍然会直接影响
    到宿主机的/root/docker-exp/busybox目录。我们需要进行容器和镜像隔离，实现在容器中进行的操作不会对镜像产生任何影响的功能
2. 原理：
    - 解压busybox.tar到/root/docker-exp/busybox目录作为容器的文件系统的只读层
    - 创建writeLayer文件夹，作为容器唯一的可写层
    - 创建mnt文件夹，作为挂载点，然后把writeLayer目录和busybox目录mount到mnt目录下
    - 最核心的代码就是 
        ```
	        dirs := "dirs=" + rootURL + "writeLayer:" + rootURL + "busybox"
	        cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntURL)
        ```
        这个代码会把我们的只读目录busybox和可写层writeLayer，以aufs的方式进行合并，并挂载到mntURL目录下
3. 运行./docker run -ti /bin/sh
4. 步骤
    - 在4.1的基础上进行，`./docker run -ti /bin/sh`
    - 我们再开一个终端，在/root/docker-exp/busybox下面可以看到有writeLayer和mnt两个目录，然后和容器里面的目录进行添加文件或者修改，可以发现在mnt会显示，而且修改项在writeLayer目录下也有
        但是busybox里面不会有变化
5. 备注：有时候会出现挂载的一些问题，如果定位不出原因，这时可以重启下机器

### 容器网络（代码不实现）
在前面，我们通过Namespace和Cgroups技术实现了容器进程的隔离，并且通过AUFS让容器拥有自己的“文件系统 ”，但是我们的容器还没有“网线”
#### 必要知识补充（极力推荐原书《自己动手实现Docker》里面的讲解）
1. linux支持创建出虚拟化的设备，可以通过虚拟化设备的组合实现多种多样的功能和网络拓扑。常见的虚拟化设备有Veth、Bridge等，这里主要介绍 构建 容器网络要用到的veth和bridge
2. Linux veth
    - veth是成对出现 的虚拟网络设备，发现到Veth一端虚拟设备的请求会在另一端的虚拟设备中发出。在容器的虚拟化场景中，经常会使用veth连接不同的网络Namespace。
        ```shell script
            # 创建两个网络Namespace
            ip netns add ns1
            ip netns add ns2
            # 创建一对Veth
            ip link add veth0 type veth peer name veth1
            # 分别将两个Veth移到两个Namespace中
            ip link set veth0 netns ns1
            ip link set veth1 nets ns2
            # 查看ns1的网络设备
            ip netns exec ns1 ip link
        ```
      当我们将请求发送到veth0这个虚拟设备时，都会原封不动地从另一个网络Namespace的网络接口veth0中出来 ，所以当我们
      给两端分别 配置不同的地址后，向虚拟网络设备的一端发送请求，就能到达这个虚拟网络设备的另一端
        ```shell script
          # 配置每个veth的网络地址和Namespace的路由
          ip netns exec ns1 ifconfig veth0 172.18.0.2/24 up
          ip netns exec ns2 ifconfig veth1 172.18.0.3/24 up
          ip netns exec ns1 route add default dev veth0
          ip nets exec ns2 route add default dev veth1
          # 通过veth一端出去的包，另外一端能够直接接收到
          ip nets exec ns1 ping -c 1 172.18.0.3
        ``` 
3. Linux Bridge 
   - Bridge 虚拟设备是用来桥接的网络设备，可以用于连接不同的网络设备，当请求到达 Bridge设备时，可以通过报文中的Mac
    地址进行广播或转发，
        ```shell script
           # 创建veth 设备并将一端移入Namespace
           ip netns add ns1
           ip link add veth0 type veth peer name veth1
           ip link set veth1 netns ns1
           # 创建网桥
           brctl addbr br0
           # 挂载网络设备
           brctl addif br0 eth0 #执行完这一步，如果是ssh连接的话，会直接失联
           brctl addif br0 veth0
        ```
4. Linux路由表
    - 路由表: 通过定义路由表来决定在某个网络Namespace中包的注射，从而定义请求会到哪个网络设备上。
        结合上面我们把宿主机的eth0给放置到网桥上，我们还需要通过路由表将网段172.18.0.0/24的请求路由到br0的网桥上；通过上面的配置当我们在ns1中进行访问宿主机的IP时，流量
        会直接通过veth1和veth0到达网桥，通过网桥到达 eth0；反之，在宿主机上ping ns1的时候，流量会通过eth0到达网桥，通过网桥上的veth0到达 veth1上，从而
        进入容器
            ```shell script
                # 启动虚拟网络设备，并设置它在Net Namespace中的IP地址
                ip link set veth0 up
                ip link set br0 up
                # 通过虚拟网络设备veth1给ns1配置网络地址
                ip netns exec ns1 ifconfig veth1 172.18.0.2/24 up
                # 设置ns1 网络空间的路由和宿主机上的路由; default 代表0.0.0.0/0，即在Net Namespace中所有流量都经过veth1的网络设备流出
                ip netns exec ns1 route add default dev veth1
                # 在宿主机上将172.18.0.0/24的网段请求路由到br0的网格上【针对宿主机的配置】
                route add -net 172.18.0.0/24 dev br0
            ```
            
5. Linux iptables
    - iptabels是对Linux内核的netfilter模块进行操作和展示的工具，用来管理包的流动和转送。iptables定义了一套链式处理的结构，在网络包传输的
        各个阶段可以使用不同的策略对包进行加工、传送或丢弃。在容器虚拟化技术中，经常使用到MASQUERADE 和DNAT，用于容器
        和宿主机外部的网络通信
    - MASQUERADE策略可以将请求包中的源地址转换成一个网络设备的地址，比如一个net namespace中网络设备的地址是172.19.0.2
        这个地址虽然 在宿主机可以路由到br0的网桥，但是到达 宿主机外部之后，是不知道如何 路由到这个IP地址的，所以如果
        请求外部地址的话，需要先通过MASQUERADE策略将这个IP地址转换成宿主机出口网上的IP；也就是说基于这个策略我们可以实现在
        Namespace中访问宿主机外的网络
        ```shell script
          # 打开IP转发
          sysctl -w net.ipv4.conf.all.forwarding=1
          #对Namespace中的包添加网络地址转换
          iptables -t nat -A POSTROUTING -s 172.18.0.0/24 -o eth0 -j MASQUERADE
        ```
    - DNAT策略也是做网络地址转换，不过它是要更换目标地址，经常用于将内部网络地址的端口映射到外部去（一般是为了在容器中给外部提供服务）
        外部应用是无法直接路由到namespace中的地址的，这时就可以使用DNAT策略
        ```shell script
          # 将宿主机上80端口的请求转发到Namespace的IP上
            iptables -t nat -A PREROUTING -p tcp -m tcp --dport 80 -j DNAT --to-destination 172.18.0.2:80
        ```
       
