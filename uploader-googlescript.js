// ^
// |
// ATTENTION: TO RUN THIS SCRIPT YOU NEED TO SAVE IT TO YOUR DRIVE FIRST, BY CLICKING THIS BUTTON

function moveUploads() {
    const uploadFolder = DriveApp.getFolderById("13EUhhaRw9ct8KhxkmnQ20q-hpm89Ynr0")
    const courseworkFolder = DriveApp.getFolderById("1ReqozOMKGLuOI1L7V51EhEWMrRthpQ5N")

    const uploadsIter = uploadFolder.getFoldersByName("courses")

    while (uploadsIter.hasNext()) {
        const coursesFolder = uploadsIter.next()
        recursiveMerge(coursesFolder, courseworkFolder)
    }
}

/**
 * @param {DriveApp.Folder} src
 * @param {DriveApp.Folder} dst
 */
function recursiveMerge(src, dst) {
    const children = src.getFolders()
    while (children.hasNext()) {
        const child = children.next()

        const matches = dst.getFoldersByName(child.getName())
        if (matches.hasNext()) {
            recursiveMerge(child, matches.next())
        } else {
            child.moveTo(dst)
        }
    }
    src.setTrashed(true)
}
