This Project is meant to be a replication of Xah Lee's computing philosophy. You can read more about it [here.](http://xahlee.info/linux/why_tiling_window_manager_sucks.html)

My reasoning for writing this in go, was because I wanted a system that would work on any UNIX based X11 environment. 
It's been tested sucessfully in GNOME4 desktop environments.

For any recruiting personal that might stumble on this while reading my resume or website, I kindly ask you to consider that I predominantely work in a Java environment so this will not be professional grade Go.

The basis of this program is to allow a user to rapidly switch between their most used applications. 

This is accomplished by binding apps to specific funtion keys that will:

a. Open an app instance binded to that key if one is not already present 

b. Switch to the last used app instance if one is already present 

c. Cycle between the app instances open if multiple exist 

The following gif will demostrate this functionality. In this example my F9 key is bounded to my web browser and F2 is bounded to my terminal.

![Demo](gif/xwm-showcase.gif)
