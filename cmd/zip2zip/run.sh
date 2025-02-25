#!/bin/sh

inputzip=./sample.d/input.zip

geninput(){
	echo generating input...

	mkdir -p sample.d

	echo hw1 > ./sample.d/hw1.txt
	echo hw2 > ./sample.d/hw2.txt
	mkdir -p ./sample.d/empty.d
	echo hw3 > ./sample.d/hw3.txt

	find \
		./sample.d \
		-mindepth 1 |
		zip \
			-@ \
			-T \
			-v \
			-o \
			"${inputzip}"
}

test -f "${inputzip}" || geninput

export ENV_INPUT_ZIP_FILENAME="${inputzip}"
export ENV_NAME_PATTERN='^sample\.d/hw[13]|.*empty'
export ENV_INCLUDE_FOUND=false
export ENV_INCLUDE_FOUND=true
export ENV_MAX_ITEM_SIZE=1024

echo ----------------------------------------------------------------
echo INPUT ZIP FILE
unzip -lv "${inputzip}"

outzip=./sample.d/out.zip

./zip2zip |
	dd \
		if=/dev/stdin \
		of="${outzip}" \
		bs=1048576 \
		status=progress

echo
echo ----------------------------------------------------------------
echo OUTPUT ZIP FILE
unzip -lv "${outzip}"
