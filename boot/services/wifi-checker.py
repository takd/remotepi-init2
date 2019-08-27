#!/usr/bin/env python
import RPi.GPIO as GPIO
import subprocess
import time

GPIO.setwarnings(False) # Ignore warning for now
GPIO.setmode(GPIO.BCM) # Use Broadcom pin numbering
GPIO.setup(10, GPIO.OUT, initial=GPIO.LOW)

while (True):
    time.sleep(1)
    ps = subprocess.Popen(['iwgetid'], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    try:
        output = subprocess.check_output(('grep', 'ESSID'), stdin=ps.stdout)
        #print(output)
        GPIO.output(10, GPIO.HIGH)
    except subprocess.CalledProcessError:
        # grep did not match any lines
        #print("No wireless networks connected")
        GPIO.output(10, GPIO.LOW)