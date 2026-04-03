graph TD
    cmd["cmd<br/>15 files"]
    internal["internal<br/>844 files"]
    tests["tests<br/>9 files"]
    cmd --> internal

    subgraph External
        chromedp(("chromedp"))
        paho_mqtt_golang(("paho.mqtt.golang"))
        proto(("proto"))
        gin(("gin"))
        v10(("v10"))
    end