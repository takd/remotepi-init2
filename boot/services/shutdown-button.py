#!/usr/bin/env python
import RPi.GPIO as GPIO
import subprocess
import os

GPIO.setmode(GPIO.BCM)
GPIO.setup(3, GPIO.IN, pull_up_down=GPIO.PUD_UP)
GPIO.wait_for_edge(3, GPIO.FALLING)

os.system("/home/pi/services/DS1302/setDateToRTC.py 1")
subprocess.call(['shutdown', '-h', 'now'], shell=False)