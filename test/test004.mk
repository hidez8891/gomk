init:
	@echo init
	@echo "" > file3.tmp

file1.tmp: file2.tmp
	@echo file1: run
	@echo "" > file1.tmp

file2.tmp: file3.tmp
	@echo file2: run
	@echo "" > file2.tmp

clean:
	@del /q file1.tmp
	@del /q file2.tmp
	@del /q file3.tmp
