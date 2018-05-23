library("jsonlite")

data <- flatten(do.call(rbind, 
                     lapply(paste(readLines(file("output-v1.json"), warn=FALSE),
                                  collapse=""), 
                            jsonlite::fromJSON)))

plot(
  x=data[,"RepoInfo.forks_count"],
  y=data[,"TotalBytes"],
  xlab="Number of forks",
  ylab="Size of documentation in bytes",
  log="xy"
)