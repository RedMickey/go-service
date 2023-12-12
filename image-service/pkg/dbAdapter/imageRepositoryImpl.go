package dbAdapter

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"image-service/pkg/core"
)

type imageRepositoryImpl struct {
	db *sql.DB
}

func NewImageRepository(db *sql.DB) core.ImageRepository {
	return &imageRepositoryImpl{
		db: db,
	}
}

func (r *imageRepositoryImpl) GetImageById(id string) (*core.ImageEntity, error) {
	imageEntity := &core.ImageEntity{}

	err := r.db.QueryRow("select * from image where id = $1", id).Scan(
		&imageEntity.Id,
		&imageEntity.Name,
		&imageEntity.Url,
		&imageEntity.CreatedDate,
		&imageEntity.UpdatedDate,
		(*pq.StringArray)(&imageEntity.AvailableFormats),
	)

	if err != nil {
		return nil, err
	}
	// defer rows.Close()

	// imageEntity := &core.ImageEntity{}

	// for rows.Next() {
	// 	err := rows.Scan(
	// 		&imageEntity.Id,
	// 		&imageEntity.Name,
	// 		&imageEntity.Url,
	// 		&imageEntity.CreatedDate,
	// 		&imageEntity.UpdatedDate,
	// 		(*pq.StringArray)(&imageEntity.AvailableFormats),
	// 	)

	// 	if err != nil {
	// 		fmt.Println(err)
	// 		continue
	// 	}

	// 	break
	// }

	return imageEntity, nil
}

func (r *imageRepositoryImpl) DeleteImageById(id string) (int, error) {
	res, err := r.db.Exec("delete from image where id = $1", id)

	if err != nil {
		return 0, err
	}

	rowsAffected, err := res.RowsAffected()

	if err != nil {
		return 0, err
	}

	fmt.Println(res)

	return int(rowsAffected), nil
}

func (r *imageRepositoryImpl) CreateImage(image core.ImageCreateDto) (*core.ImageEntity, error) {
	imageEntity := &core.ImageEntity{}

	err := r.db.QueryRow(
		"insert into image(id, name, url, \"availableFormats\") values($1, $2, $3, $4) returning *",
		image.Id,
		image.Name,
		image.Url,
		pq.Array(image.AvailableFormats),
	).Scan(
		&imageEntity.Id,
		&imageEntity.Name,
		&imageEntity.Url,
		&imageEntity.CreatedDate,
		&imageEntity.UpdatedDate,
		(*pq.StringArray)(&imageEntity.AvailableFormats),
	)

	if err != nil {
		return nil, err
	}

	return imageEntity, nil
}

func (r *imageRepositoryImpl) UpdateImage(image core.ImageEntity) (*core.ImageEntity, error) {
	imageEntity := &core.ImageEntity{}

	err := r.db.QueryRow(
		"update image set name = $1, url = $2, \"updatedDate\" = $3, \"availableFormats\" = $4 where id = $5 returning *",
		image.Name,
		image.Url,
		image.UpdatedDate,
		pq.Array(image.AvailableFormats),
		image.Id,
	).Scan(
		&imageEntity.Id,
		&imageEntity.Name,
		&imageEntity.Url,
		&imageEntity.CreatedDate,
		&imageEntity.UpdatedDate,
		(*pq.StringArray)(&imageEntity.AvailableFormats),
	)

	if err != nil {
		return nil, err
	}

	return imageEntity, nil
}
