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

### Union File System
1. 简称是UnionFS，一种可以把其他文件系统联合到一个联合挂载点的文件系统服务，当我们对这个联合文件系统进行写操作的时候，系统是真正写到了一个新的文件，
这样看下来这个虚拟的联合文件系统是可以对任何文件进行操作的，但是其实它并没有改变原来的文件，这其实就是“写时复制”，一种资源管理技术；它的思想是
   如果一个资源是重复的，但没有任何修改，这时并不需要立即创建一个新的资源，这个资源 可以被新旧实例共享。创建新资源发生在第一次写操作时。通过这种方式，
   可以显著减少对未修改资源复制所带来的消耗。
   
2. Docker的存储方式有AUFS、overlayFS等。默认的是overlay2

### 构造容器
#### 构造实现run命令版本的容器分支 3.1
0. 目标：运行./docker run -ti /bin/sh; ./docker run -ti /bin/ls
1. 对着书把代码里面的注释补充下
2. 构建go build -o mydocker .
3. 遇到的问题： github.com/xianlubird/mydocker/issues/33;
4. 调试：dlv exec ./docker -- run -ti /bin/bash
5. 大致的流程：解析到run的时候会调用runCommand里面的action，这个action会调用根据参数调用initCommand，初始化容器资源，运行输入的命令/bin/sh