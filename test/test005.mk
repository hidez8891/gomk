all: echo1 echo2

echo1: echo3
	@echo echo1

echo2: echo4
	@echo echo2

echo3: echo4
	@echo echo3

echo4:
	@echo echo4
