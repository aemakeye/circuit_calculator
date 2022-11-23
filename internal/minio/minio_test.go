package minio

//https://github.com/minio/minio-go/tree/master/examples/s3

import (
	"bytes"
	"context"
	"github.com/aemakeye/circuit_calculator/internal/calculator"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"io/ioutil"
	"strconv"
	"testing"
)

func TestNewMinioStorage(t *testing.T) {
	//var endpoint = "minio-console.local.net:9000"
	var endpoint = "localhost:9000"
	var user = "calculator"
	var password = "c@1cu1@t0r"
	var bucket = "calculator"
	var diagram = []byte(`
			<mxfile host="65bd71144e">
				<diagram id="uweCVhkyVy6MirBnUyNJ" name="Page-1">
				<mxGraphModel dx="354" dy="159" grid="1" gridSize="10" guides="1" tooltips="1" connect="0" arrows="1" fold="1" page="1" pageScale="1" pageWidth="827" pageHeight="1169" math="0" shadow="0">
					<root>
						<mxCell id="0"/>
						<mxCell id="1" parent="0"/>
						<mxCell id="3" value="" style="pointerEvents=1;verticalLabelPosition=bottom;shadow=0;dashed=0;align=center;html=1;verticalAlign=top;shape=mxgraph.electrical.resistors.resistor_1;" vertex="1" parent="1">
							<mxGeometry x="110" y="140" width="100" height="20" as="geometry"/>
						</mxCell>
						<mxCell id="4" value="" style="pointerEvents=1;verticalLabelPosition=bottom;shadow=0;dashed=0;align=center;html=1;verticalAlign=top;shape=mxgraph.electrical.inductors.inductor_3;" vertex="1" parent="1">
							<mxGeometry x="260" y="120" width="100" height="10" as="geometry"/>
						</mxCell>
						<mxCell id="6" value="" style="pointerEvents=1;verticalLabelPosition=bottom;shadow=0;dashed=0;align=center;html=1;verticalAlign=top;shape=mxgraph.electrical.capacitors.capacitor_1;" vertex="1" parent="1">
							<mxGeometry x="280" y="170" width="100" height="60" as="geometry"/>
						</mxCell>
						<mxCell id="7" value="" style="endArrow=none;html=1;exitX=0.993;exitY=0.505;exitDx=0;exitDy=0;exitPerimeter=0;entryX=0.004;entryY=0.507;entryDx=0;entryDy=0;entryPerimeter=0;" edge="1" parent="1" source="3" target="6">
							<mxGeometry width="50" height="50" relative="1" as="geometry">
								<mxPoint x="240" y="200" as="sourcePoint"/>
								<mxPoint x="290" y="150" as="targetPoint"/>
							</mxGeometry>
						</mxCell>
						<mxCell id="8" value="" style="endArrow=none;html=1;entryX=1.002;entryY=1.052;entryDx=0;entryDy=0;entryPerimeter=0;exitX=0.998;exitY=0.507;exitDx=0;exitDy=0;exitPerimeter=0;edgeStyle=elbowEdgeStyle;" edge="1" parent="1" source="6" target="4">
							<mxGeometry width="50" height="50" relative="1" as="geometry">
								<mxPoint x="240" y="180" as="sourcePoint"/>
								<mxPoint x="290" y="130" as="targetPoint"/>
								<Array as="points">
									<mxPoint x="420" y="150"/>
									<mxPoint x="390" y="170"/>
								</Array>
							</mxGeometry>
						</mxCell>
						<mxCell id="9" value="" style="endArrow=none;html=1;exitX=-0.002;exitY=1.028;exitDx=0;exitDy=0;exitPerimeter=0;" edge="1" parent="1" source="4">
							<mxGeometry width="50" height="50" relative="1" as="geometry">
								<mxPoint x="210" y="125" as="sourcePoint"/>
								<mxPoint x="210" y="150" as="targetPoint"/>
							</mxGeometry>
						</mxCell>
						<mxCell id="10" value="" style="endArrow=classic;html=1;exitX=0.012;exitY=0.545;exitDx=0;exitDy=0;exitPerimeter=0;entryX=0.015;entryY=0.513;entryDx=0;entryDy=0;entryPerimeter=0;edgeStyle=orthogonalEdgeStyle;" edge="1" parent="1" source="3" target="6">
							<mxGeometry width="50" height="50" relative="1" as="geometry">
								<mxPoint x="240" y="210" as="sourcePoint"/>
								<mxPoint x="290" y="160" as="targetPoint"/>
								<Array as="points">
									<mxPoint x="111" y="201"/>
								</Array>
							</mxGeometry>
						</mxCell>
					</root>
				</mxGraphModel>
			</diagram>
		</mxfile>
		`)

	m, err := NewMinioStorage(zap.NewNop(), endpoint, bucket, user, password, false)

	assert.NoError(t, err)

	t.Run("test minio upload", func(t *testing.T) {

		err := m.UploadTextFile(context.Background(), zap.NewNop(), bytes.NewReader(diagram), "test/diagram.xml")

		assert.NoError(t, err)
		//t.Logf("VersionID: %s", info.VersionID)
		//t.Logf("ETAG: %s", info.ETag)
		info2, err := m.Client.ListBuckets(context.Background())

		assert.NoError(t, err)

		t.Logf("%s", info2[0].Name)

	})
	t.Run("learn getting object", func(t *testing.T) {
		objReader, err := m.Client.GetObject(
			context.Background(),
			bucket,
			"test-diagram.xml",
			minio.GetObjectOptions{
				ServerSideEncryption: nil,
				VersionID:            "",
				PartNumber:           0,
				Checksum:             false,
				Internal:             minio.AdvancedGetOptions{},
			},
		)
		assert.NoError(t, err)
		defer objReader.Close()

		buf, err := ioutil.ReadAll(objReader)
		assert.NoError(t, err)

		t.Logf("%s", buf)
	})

	t.Run("learn list objects", func(t *testing.T) {

		for i := 0; i < 3; i++ {
			info, err := m.Client.PutObject(
				context.Background(),
				bucket,
				"test-diagram"+strconv.Itoa(i)+".xml",
				bytes.NewReader(diagram),
				int64(len(diagram)),
				minio.PutObjectOptions{ContentType: "application/octet-stream"},
			)
			assert.NoError(t, err)
			t.Logf("uploaded i: %d", i)
			t.Logf("VersionID: %s", info.VersionID)
			t.Logf("ETAG: %s", info.ETag)
		}

		infoChan := m.Ls(context.Background())

		for obj := range infoChan {
			assert.NotEmpty(t, obj)
			t.Logf("obj info: %s", obj.Name)
		}

	})

	t.Run("load version", func(t *testing.T) {
		infoChan := m.LsVersions(context.Background(), &calculator.Diagram{
			UUID:     "",
			Body:     "",
			Error:    "",
			Items:    nil,
			Name:     "test-diagram0.xml",
			Versions: nil,
		})

		for obj := range infoChan {
			t.Logf("test_diagram0.xml version found: %s", obj.Version)

		}

	})
	t.Run("try get object, version latest, OK", func(t *testing.T) {
		obj, err := m.LoadDiagramByName(
			context.Background(),
			zap.NewNop(),
			"test-diagram.xml",
			"",
		)
		assert.NoError(t, err)
		objReader := bytes.NewReader(obj)

		buf, err := ioutil.ReadAll(objReader)
		assert.NoError(t, err)

		t.Logf("%s", buf)
	})
	t.Run("try get object, version latest, FAIL", func(t *testing.T) {
		_, err := m.LoadDiagramByName(
			context.Background(),
			zap.NewNop(),
			"test-diagram-FAIL.xml",
			"",
		)
		assert.Error(t, err)
		t.Logf("%s", err)
	})
}
