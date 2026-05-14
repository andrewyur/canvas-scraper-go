# canvas-scraper-go

## Directions

1. download the relevant executable for your platform from the releases section
2. get the value of the `canvas_session` cookie from your web browser
3. open up a terminal, and run the executable from the command line
4. paste in the text you copied in step 2. when the program asks for it
5. after the program finishes running, upload the `courses` folder to the file repository, inside Coursework -> Upload
6. Run the google script to merge the files into the coursework repo

## Maintenance

This piece of software is heavily dependent on the canvas API remaining stable and the naming conventions of the courses 
staying the same, so the moment one of these things changes someone/something will likely have to go in and fix this. 
I primarily used [the public canvas API documentation](https://lms.au.af.edu/doc/api/), which will likely still be accurate for a while. if the 
api gets changed/updated there may be a chance you can get some of the bits and pieces you might need from reverse engineering
the api used by the canvas website, which can be done by watching the network requests in your browser when you visit the webpage.

### Development Priorities

1. Make it as simple/quick as possible for the end user to upload their canvas coursework to the drive
    - no uncommon runtime dependencies
    - make binary as portable as possible within reason
2. Make the code easy to fix for lodgers down the line for when the uploader eventually breaks
    - keep the code easy to compile & distribute
    - keep everything simple and easily readable