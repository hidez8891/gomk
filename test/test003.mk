# loop rules

loop: rule1

rule1: rule2
	@echo "rule1"

rule2: rule3
	@echo "rule2"

rule3: rule1
	@echo "rule3"
