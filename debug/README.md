
// 服务开启StartPprofSrv，监听地址例如： http://HOST:PORT
// 通过浏览器访问 http://HOST:PORT/debug/pprof/ 可以查看部分profile状态参数
// 使用go tool  pprof 工具进行分析
// 内存 go tool pprof http://HOST:PORT/debug/pprof/heap
// CPU  go tool pprof http://HOST:PORT/debug/pprof/profile
//      go tool pprof http://HOST:PORT/debug/pprof/block
// go tool pprof ./binary ./heap
// $ go tool pprof -http=":8081" ./binary ./heap  火焰图
// 通过 http://HOST:PORT/debug/pprof/trace 获取TRACEFILE文件
// 使用go tool trace  TRACEFILE 在浏览器中查看性能分析结果
// windows安装graphviz， https://graphviz.gitlab.io/_pages/Download/Download_windows.html 下载安装
//  安装完成后，把安装路径bin目录 例如默认值C:\Program Files (x86)\Graphviz2.38\bin  加入Path环境变量中
// linux 安装  yum -y install graphviz      sudo apt-get install graphviz
