.PHONY: install
install:
	rm -rf bin
	GOBIN=$(PWD)/bin go install ./...

.PHONY: run-both
run-both:
	go run cmd/SurfstoreServerExec/main.go -s both -p 8081 -l localhost:8081

.PHONY: run-blockstore
run-blockstore:
	go run cmd/SurfstoreServerExec/main.go -s block -p 8081 -l

.PHONY: run-metastore
run-metastore:
	go run cmd/SurfstoreServerExec/main.go -s meta -l localhost:8081

# 串行测试

# client1 空文件夹
# 预期 index.txt, 内容为空
emptyFile:
	rm -rf test/emptyFile
	mkdir test/emptyFile
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/emptyFile 10240

# client1 提交一张猫
# 预期：index.txt 内容为一张猫
singleClient:
	rm -rf test/singleClient
	mkdir test/singleClient
	cp statics/kitten.jpeg test/singleClient/kitten.jpeg
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/singleClient 10240

# client1 sync提交一张猫和article之后，client2 sync
# 预期：client2 存在双内容
twoClient1:
	rm -rf test/twoClient1
	mkdir test/twoClient1 test/twoClient1/client1 test/twoClient1/client2
	cp statics/kitten.jpeg test/twoClient1/client1/kitten.jpeg
	cp statics/article.txt test/twoClient1/client1/article.txt
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/twoClient1/client1 10240
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/twoClient1/client2 10240

# client1 sync成猫，client1更新为狗再次sync
# 预期：client1内容被更新成狗,版本号为2
updateCommit1:
	rm -rf test/updateCommit1
	mkdir test/updateCommit1 test/updateCommit1/client1
	cp statics/kitten.jpeg  test/updateCommit1/client1/image.jpeg
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/updateCommit1/client1 10240
	cp statics/dog.jpeg  test/updateCommit1/client1/image.jpeg
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/updateCommit1/client1 10240

# client1 sync成猫，client1更新为猫再次sync
# 预期：client1内容还是猫,版本号为1
updateCommit2:
	rm -rf test/updateCommit2
	mkdir test/updateCommit2 test/updateCommit2/client1
	cp statics/kitten.jpeg  test/updateCommit2/client1/image.jpeg
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/updateCommit2/client1 10240
	cp statics/kitten.jpeg  test/updateCommit2/client1/image.jpeg
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/updateCommit2/client1 10240

# client1 client2 sync成猫，client2更新为狗再次sync， client1 sync
# 预期：client1内容被更新成狗,版本号为2
updateCommit3:
	rm -rf test/updateCommit3
	mkdir test/updateCommit3 test/updateCommit3/client1 test/updateCommit3/client2
	cp statics/kitten.jpeg  test/updateCommit3/client1/image.jpeg
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/updateCommit3/client1 10240
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/updateCommit3/client2 10240
	cp statics/dog.jpeg  test/updateCommit3/client2/image.jpeg
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/updateCommit3/client2 10240
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/updateCommit3/client1 10240

# client1 client2 sync成猫，client2更新为狗，client1更新为鱼，client1 sync,client2 sync
# 预期：client2内容狗被抛弃更新成鱼,版本号为2
updateCommit4:
	rm -rf test/updateCommit4
	mkdir test/updateCommit4 test/updateCommit4/client1 test/updateCommit4/client2
	cp statics/kitten.jpeg  test/updateCommit4/client1/image.jpeg
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/updateCommit4/client1 10240
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/updateCommit4/client2 10240
	cp statics/dog.jpeg  test/updateCommit4/client2/image.jpeg
	cp statics/fish.jpeg  test/updateCommit4/client1/image.jpeg
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/updateCommit4/client1 10240
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/updateCommit4/client2 10240

# client1 client2 原有猫和狗， client1 sync 然后client2 sync
# 预期：client2狗内容被替换成猫
noIndexCommit1:
	rm -rf test/noIndexCommit1
	mkdir test/noIndexCommit1 test/noIndexCommit1/client1 test/noIndexCommit1/client2
	cp statics/kitten.jpeg  test/noIndexCommit1/client1/image.jpeg
	cp statics/dog.jpeg  test/noIndexCommit1/client2/image.jpeg
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/noIndexCommit1/client1 10240
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/noIndexCommit1/client2 10240

# client1 原有猫sync后删除再sync
# 预期：client1版本为2， hashList为“0”
deleteCommit1:
	rm -rf test/deleteCommit1
	mkdir test/deleteCommit1
	cp statics/kitten.jpeg  test/deleteCommit1/image.jpeg
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/deleteCommit1 10240
	rm test/deleteCommit1/image.jpeg
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/deleteCommit1 10240

# client1 存入空文件后sync， client2 sync后删除再sync
# 预期：client1存在空文件，版本号为1，client2不存在，版本号为2
emptyFileCommit1:
	rm -rf test/emptyFileCommit1
	mkdir test/emptyFileCommit1  test/emptyFileCommit1/client1 test/emptyFileCommit1/client2
	cp statics/nothing test/emptyFileCommit1/client1/nothing
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/emptyFileCommit1/client1 10240
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/emptyFileCommit1/client2 10240
	rm test/emptyFileCommit1/client2/nothing
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/emptyFileCommit1/client2 10240


# 并行测试

# client1 sync上传五个文件的同时client2 并行sync
# 预期： 根据睡眠时间不同，结果不同
syncTwoClient1:
	rm -rf test/syncTwoClient1
	mkdir test/syncTwoClient1 test/syncTwoClient1/client1 test/syncTwoClient1/client2
	cp statics/kitten.jpeg test/syncTwoClient1/client1/kitten.jpeg
	cp statics/fish.jpeg test/syncTwoClient1/client1/fish.jpeg
	cp statics/dog.jpeg test/syncTwoClient1/client1/dog.jpeg
	cp statics/article.txt test/syncTwoClient1/client1/article.txt
	cp statics/nothing test/syncTwoClient1/client1/nothing
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/syncTwoClient1/client1 10240 &
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/syncTwoClient1/client2 10240

# client1 有猫和文章，client2有空文件和狗，并行commit
# 预期 两个文件中的image必须相同
syncConflict1:
	rm -rf test/syncConflict1
	mkdir test/syncConflict1 test/syncConflict1/client1 test/syncConflict1/client2
	cp statics/kitten.jpeg test/syncConflict1/client1/image.jpeg
	cp statics/article.txt test/syncConflict1/client1/article.txt
	cp statics/dog.jpeg test/syncConflict1/client2/image.jpeg
	cp statics/nothing test/syncConflict1/client2/nothing
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/syncConflict1/client1 10240 &
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/syncConflict1/client2 10240

# 6个文件并行上传
syncConflict2:
	rm -rf test/syncConflict2
	mkdir test/syncConflict2 test/syncConflict2/client1 test/syncConflict2/client2
	cp statics/kitten.jpeg test/syncConflict2/client1/image1.jpeg
	cp statics/fish.jpeg test/syncConflict2/client1/image2.jpeg
	cp statics/dog.jpeg test/syncConflict2/client1/image3.jpeg
	cp statics/kitten.jpeg test/syncConflict2/client1/image4.jpeg
	cp statics/fish.jpeg test/syncConflict2/client1/image5.jpeg
	cp statics/dog.jpeg test/syncConflict2/client1/image6.jpeg
	cp statics/dog.jpeg test/syncConflict2/client2/image1.jpeg
	cp statics/kitten.jpeg test/syncConflict2/client2/image2.jpeg
	cp statics/fish.jpeg test/syncConflict2/client2/image3.jpeg
	cp statics/dog.jpeg test/syncConflict2/client2/image4.jpeg
	cp statics/kitten.jpeg test/syncConflict2/client2/image5.jpeg
	cp statics/fish.jpeg test/syncConflict2/client2/image6.jpeg
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/syncConflict2/client1 40 &
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/syncConflict2/client2 40

# 5个客户大文件 压力测试
# 你自己先生成五个随机大文件得
# 预期 完全一致！
syncConflict3:
	rm -rf test/syncConflict3
	mkdir test/syncConflict3 test/syncConflict3/client1 test/syncConflict3/client2 test/syncConflict3/client3 test/syncConflict3/client4 test/syncConflict3/client5
	cp statics/1 test/syncConflict3/client1/file1
	cp statics/2 test/syncConflict3/client1/file2
	cp statics/3 test/syncConflict3/client1/file3
	cp statics/2 test/syncConflict3/client2/file1
	cp statics/3 test/syncConflict3/client2/file2
	cp statics/4 test/syncConflict3/client2/file3
	cp statics/3 test/syncConflict3/client3/file1
	cp statics/4 test/syncConflict3/client3/file2
	cp statics/5 test/syncConflict3/client3/file3
	cp statics/4 test/syncConflict3/client4/file1
	cp statics/5 test/syncConflict3/client4/file2
	cp statics/1 test/syncConflict3/client4/file3
	cp statics/5 test/syncConflict3/client5/file1
	cp statics/1 test/syncConflict3/client5/file2
	cp statics/2 test/syncConflict3/client5/file3
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/syncConflict3/client1 120000 &
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/syncConflict3/client2 120000 &
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/syncConflict3/client3 120000 &
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/syncConflict3/client4 120000 &
	go run cmd/SurfstoreClientExec/main.go -d localhost:8081 test/syncConflict3/client5 120000 
	head -c 10 test/syncConflict3/client1/file1