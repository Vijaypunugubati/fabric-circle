pipeline {
    agent any

    options {
        timestamps ()
    }

    environment {
        ARCH="amd64"
        GOVER="1.11.5"
        GOROOT="${env.WORKSPACE}/go${env.GOVER}/go"
        GOPATH="${env.WORKSPACE}/go"
        PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:${env.GOROOT}/bin:${env.GOPATH}/bin"
    }

     stages {
        stage('Setup Environment') {
            steps {
                cleanWs deleteDirs: true
                sh label: "Create GOROOT Directory", script: "mkdir -p go${GOVER}"
                sh label: "Create GOPATH Directory", script: "mkdir -p go/src/github.com/hyperledger/fabric"
                sh label: "Download Go ${env.GOVER}", script: "wget -nv https://dl.google.com/go/go${GOVER}.linux-${ARCH}.tar.gz"
                sh label: "Extract Go ${env.GOVER}", script: "tar -C ./go${GOVER} -xzf go${GOVER}.linux-${ARCH}.tar.gz"
                sh label: "Remove Go ${env.GOVER} tarball", script: "rm -rf go${GOVER}.linux-${ARCH}.tar.gz"
                dir('go/src/github.com/hyperledger/fabric') {
                    checkout scm
                }
            } 
        }

        stage('Run Tests') {
            parallel {
                stage('Run Unit Tests') {
                    when {
                        branch 'PR-*'
                    }
                    steps {
                        dir('go/src/github.com/hyperledger/fabric') {
                            sh label: 'Running Fabric Unit Tests', script: "make unit-tests"
                        }
                    }
                }

                stage('Run Integration Tests') {
                    when {
                        branch 'PR-*'
                    }
                    steps {
                        dir('go/src/github.com/hyperledger/fabric') {
                            sh label: 'Running Fabric Integration Tests', script: "make integration-tests"
                        }
                    }
                }
            }
        }

        stage('Promote to QA') {
            when {
                branch 'dev'
            }
            steps {
                lock(resource: "lock_${env.NODE_NAME}_${env.BRANCH_NAME}", inversePrecedence: true) {
                    dir('go/src/github.com/hyperledger/fabric') {
                        sh label: 'Promote to QA', script: "git checkout qa; git merge origin/dev; git push origin qa;"
                    }
                }
            }
        }

        stage('Verify Merge on QA') {
            parallel {
                stage('Run Unit Tests') {
                    when {
                        branch 'qa'
                    }
                    steps {
                        dir('go/src/github.com/hyperledger/fabric') {
                            sh label: 'Running Fabric Unit Tests', script: "make unit-tests"
                        }
                    }
                }

                stage('Run Integration Tests') {
                    when {
                        branch 'qa'
                    }
                    steps {
                        dir('go/src/github.com/hyperledger/fabric') {
                            sh label: 'Running Fabric Integration Tests', script: "make integration-tests"
                        }
                    }
                }
            }
        }

        stage('Promote to Master') {
            when {
                branch 'qa'
            }
            steps {
                lock(resource: "lock_${env.NODE_NAME}_${env.BRANCH_NAME}", inversePrecedence: true) {
                    dir('go/src/github.com/hyperledger/fabric') {
                        sh label: 'Promote to Master', script: "git checkout master; git merge origin/qa; git push origin master;"
                    }
                }
            }
        }

        stage('Verify Merge on Master') {
            parallel {
                stage('Run Unit Tests') {
                    when {
                        branch 'master'
                    }
                    steps {
                        dir('go/src/github.com/hyperledger/fabric') {
                            sh label: 'Running Fabric Unit Tests', script: "make unit-tests"
                        }
                    }
                }

                stage('Run Integration Tests') {
                    when {
                        branch 'master'
                    }
                    steps {
                        dir('go/src/github.com/hyperledger/fabric') {
                            sh label: 'Running Fabric Integration Tests', script: "make integration-tests"
                        }
                    }
                }
            }
        }
    }
}
