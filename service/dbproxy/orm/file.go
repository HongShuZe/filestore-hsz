package orm

import (
	"database/sql"
	mydb "filestore-hsz/service/dbproxy/conn"
	"fmt"
	"log"
)

// 文件上传完成， 保存meta
func OnFileUploadFinished(filehash string, filename string,
	filesize int64, fileaddr string) (res ExecResult) {
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_file (`file_sha1`, `file_name`, `file_size`, " +
			"`file_addr`, `status`) values (?,?,?,?,1)")
	if err != nil {
		log.Println("Failed to prepare prepare statement., err" + err.Error())
		res.Suc = false
		return
	}
	defer stmt.Close()

	ret, err := stmt.Exec(filehash, filename, filesize, fileaddr)
	if err != nil {
		log.Println(err.Error())
		res.Suc = false
		return
	}
	if rf, err := ret.RowsAffected(); nil == err {
		if rf <= 0 {
			log.Printf("file with hash:%s has been upload before", filehash)
		}
		res.Suc = true
		return
	}
	res.Suc = false
	return
}
// 从mysql获取元文件信息
func GetFileMeta(filehash string) (res ExecResult) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_sha1, file_addr, file_name, file_size from tbl_file " +
			"where file_sha1=? and status=1 limit 1")
	if err != nil {
		log.Println(err.Error())
		res.Suc = false
		res.Msg = err.Error()
		return
	}
	defer stmt.Close()

	tfile := TableFile{}
	err = stmt.QueryRow(filehash).Scan(
		&tfile.FileHash, &tfile.FileAddr, &tfile.FileName, &tfile.FileSize)
	if err != nil {
		if err == sql.ErrNoRows {
			// 查不到对应记录， 返回参数及错误均为nil
			res.Suc = true
			res.Data = nil
			return
		} else {
			log.Println(err.Error())
			res.Suc = false
			res.Msg = err.Error()
			return
		}
	}
	res.Suc = true
	res.Data = tfile
	return
}

/*// 文件是否已经上传
func IsFileUploaded(filehash string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"select 1 from tbl_file where file_sha1=? and status=1 limit 1")

	rows, err := stmt.Query(filehash)
	if err != nil {
		return false
	}else if rows == nil || !rows.Next() {
		return false
	}
	return true
}*/

// 从mysql批量获取文件元信息
func GetFileMetaList(limit int64) (res ExecResult) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_sha1, file_addr, file_name, file_size from tbl_file" +
			"where status=1 limit ?")
	if err != nil {
		log.Println(err.Error())
		res.Suc = false
		res.Msg = err.Error()
		return
	}
	defer stmt.Close()

	rows, err := stmt.Query(limit)
	if err != nil {
		log.Println(err.Error())
		res.Suc = false
		res.Msg = err.Error()
		return
	}

	columns, _ := rows.Columns()
	values := make([]sql.RawBytes, len(columns))
	var tfiles []TableFile
	for i := 0; i < len(values) && rows.Next(); i++ {
		tfile := TableFile{}
		err = rows.Scan(&tfile.FileHash, &tfile.FileAddr,
			&tfile.FileName, &tfile.FileSize)

		if err != nil {
			log.Println(err.Error())
			break
		}
		tfiles = append(tfiles, tfile)
	}
	fmt.Println(len(tfiles))
	res.Suc = true
	res.Data = tfiles
	return
}

/*// 文件删除（标记删除， 即把status改为2）
func OnFileRemoved(filehash string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"update tbl_file set status=2 where file_sha1=? and status=1 limit 1")
	if err != nil {
		fmt.Println("Failed to prepare statement, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(filehash)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if rf, err := ret.RowsAffected(); nil == err {
		if rf <= 0 {
			fmt.Printf("File with hash:%s not upload", filehash)
		}
		return true
	}

	return false
}*/

// 更新文件的存储地址(如文件被转移了)
func UpdateFileLocation(filehash string, fileaddr string) (res ExecResult) {
	stmt, err := mydb.DBConn().Prepare(
		"update tbl_file set `file_addr`=? where `file_sha1`=? limit 1")
	if err != nil {
		log.Println("预编译sql失败, err:" + err.Error())
		res.Suc = false
		res.Msg = err.Error()
		return
	}

	ret, err := stmt.Exec(fileaddr, filehash)
	if err != nil {
		log.Println(err.Error())
		res.Suc = false
		res.Msg = err.Error()
		return
	}
	if rf, err := ret.RowsAffected(); nil == err {
		if rf <= 0 {
			log.Printf("更新文件localtion失败, filehash:%s", filehash)
			res.Suc = false
			res.Msg = "无记录更新"
			return
		}
		res.Suc = true
		return
	} else {
		res.Suc = false
		res.Msg = err.Error()
		return
	}
}