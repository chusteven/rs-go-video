import subprocess
import re
import json


def list_displays():
    output = subprocess.check_output("xrandr --listmonitors", shell=True).decode(
        "utf-8"
    )
    displays = []
    screen_id = 1

    monitor_info = re.findall(
        r" (\d+): \+(\S+) (\d+)/(\d+)x(\d+)/(\d+) \+(\d+)\+(\d+)", output
    )
    for idx, match in enumerate(monitor_info):
        width = float(match[2])
        height = float(match[3])
        origin_x = float(match[5])
        origin_y = float(match[6])
        displays.append(
            {
                "screen_id": screen_id,
                "width": width,
                "height": height,
                "origin_x": origin_x,
                "origin_y": origin_y,
                "index": idx,
            }
        )
        screen_id += 1
    return displays


if __name__ == "__main__":
    displays = list_displays()
    print(json.dumps(displays))
