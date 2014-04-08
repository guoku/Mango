package jobs

import (
    "Mango/gojobs/log"
    "labix.org/v2/mgo"
)

func MongoInit(host, db, collection string) *mgo.Collection {
    session, err := mgo.Dial(host)
    if err != nil {
        log.Error(err)
        panic(err)
    }
    return session.DB(db).C(collection)
}
