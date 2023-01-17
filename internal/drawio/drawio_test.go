package drawio

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"reflect"
	"testing"
)

func TestXMLBasic(t *testing.T) {
	var diagramBody = []byte(`
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

	t.Run("test import", func(t *testing.T) {
		//logger := zap.NewNop()
		D := &Mxfile{}
		err := xml.Unmarshal(diagramBody, D)
		assert.NoError(t, err)
	})
	t.Run("get diagramBody id and check type is int", func(t *testing.T) {
		D := &Mxfile{}
		_ = xml.Unmarshal(diagramBody, D)
		assert.Equal(t, "uweCVhkyVy6MirBnUyNJ", D.Diagram.Id)
		assert.IsType(t, reflect.TypeOf(0), reflect.TypeOf(D.Diagram.MxGraphModel.Root.MxCells[3].Id))
	})
}

func TestNewItemDTO(t *testing.T) {
	tests := []struct {
		name           string
		xmlin          []byte
		expectedResult []string
		expectedError  string
	}{
		{
			"capacitor",
			[]byte(`
							<mxfile host="65bd71144e">
								<diagram id="QjKBXMU_Vo2TtaLlkMbm" name="Page-1">
									<mxGraphModel dx="1718" dy="484" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="827" pageHeight="1169" math="0" shadow="0">
										<root>
											<mxCell id="0"/>
											<mxCell id="1" parent="0"/>
                								<mxCell id="12" value="" style="pointerEvents=1;verticalLabelPosition=bottom;shadow=0;dashed=0;align=center;html=1;verticalAlign=top;shape=mxgraph.electrical.capacitors.ganged_capacitor;rotation=60;" vertex="1" parent="1">
													<mxGeometry x="360" y="200" width="100" height="130" as="geometry"/>
											</mxCell>
										</root>
									</mxGraphModel>
								</diagram>
							</mxfile>
				`),
			[]string{"capacitors", "ganged_capacitor"},
			"",
		},
		{
			"resistor",
			[]byte(`
							<mxfile host="65bd71144e">
								<diagram id="QjKBXMU_Vo2TtaLlkMbm" name="Page-1">
									<mxGraphModel dx="1718" dy="484" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="827" pageHeight="1169" math="0" shadow="0">
										<root>
											<mxCell id="0"/>
											<mxCell id="1" parent="0"/>
											<mxCell id="13" value="" style="pointerEvents=1;verticalLabelPosition=bottom;shadow=0;dashed=0;align=center;html=1;verticalAlign=top;shape=mxgraph.electrical.resistors.memristor_1;" vertex="1" parent="1">
												<mxGeometry x="360" y="250" width="100" height="20" as="geometry"/>
											</mxCell>
										</root>
									</mxGraphModel>
								</diagram>
							</mxfile>
				`),
			[]string{"resistors", "memristor_1"},
			"",
		},
		{
			"inductor",
			[]byte(`
							<mxfile host="65bd71144e">
								<diagram id="QjKBXMU_Vo2TtaLlkMbm" name="Page-1">
									<mxGraphModel dx="1718" dy="484" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="827" pageHeight="1169" math="0" shadow="0">
										<root>
											<mxCell id="0"/>
											<mxCell id="1" parent="0"/>
											<mxCell id="14" value="" style="pointerEvents=1;verticalLabelPosition=bottom;shadow=0;dashed=0;align=center;html=1;verticalAlign=top;shape=mxgraph.electrical.inductors.saturating_transformer;" vertex="1" parent="1">
												<mxGeometry x="310" y="190" width="200" height="150" as="geometry"/>
											</mxCell>
										</root>
									</mxGraphModel>
								</diagram>
							</mxfile>
				`),
			[]string{"inductors", "saturating_transformer"},
			"",
		},
		{
			"line is OK",
			[]byte(`<mxfile host="65bd71144e">
					<diagram id="uweCVhkyVy6MirBnUyNJ" name="Page-1">
						<mxGraphModel dx="1718" dy="484" grid="1" gridSize="10" guides="1" tooltips="1" connect="0" arrows="1" fold="1" page="1" pageScale="1" pageWidth="827" pageHeight="1169" math="0" shadow="0">
							<root>
								<mxCell id="0"/>
								<mxCell id="1" parent="0"/>
								<mxCell id="13" value="" style="endArrow=none;html=1;exitX=-0.04;exitY=0.6;exitDx=0;exitDy=0;exitPerimeter=0;entryX=1;entryY=0.55;entryDx=0;entryDy=0;entryPerimeter=0;" edge="1" parent="1" source="14" target="16">
									<mxGeometry width="50" height="50" relative="1" as="geometry">
										<mxPoint x="370" y="420" as="sourcePoint"/>
										<mxPoint x="330" y="290" as="targetPoint"/>
									</mxGeometry>
								</mxCell>
								<mxCell id="14" value="" style="pointerEvents=1;verticalLabelPosition=bottom;shadow=0;dashed=0;align=center;html=1;verticalAlign=top;shape=mxgraph.electrical.resistors.resistor_1;" vertex="1" parent="1">
									<mxGeometry x="460" y="180" width="100" height="20" as="geometry"/>
								</mxCell>
								<mxCell id="16" value="" style="pointerEvents=1;verticalLabelPosition=bottom;shadow=0;dashed=0;align=center;html=1;verticalAlign=top;shape=mxgraph.electrical.resistors.resistor_1;" vertex="1" parent="1">
									<mxGeometry x="130" y="300" width="100" height="20" as="geometry"/>
								</mxCell>
							</root>
						</mxGraphModel>
					</diagram>
				</mxfile>`),
			[]string{"lines", "line"},
			"",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			D := &Mxfile{}
			_ = xml.Unmarshal(test.xmlin, D)
			var elem MxCell
			elem = D.Diagram.MxGraphModel.Root.MxCells[2]
			dto := NewItemDTO(&elem, "ijifjvifjv")
			assert.Equal(t, dto.Class, test.expectedResult[0])
		})
	}
}

func TestController_XmlToItems_Errors(t *testing.T) {
	tests := []struct {
		name           string
		document       []byte
		expectedResult []string
		expectedError  error
		prepare        func(t *testing.T) string
	}{
		{
			"no id in document",
			[]byte(`bad document`),
			nil,
			fmt.Errorf("%s", "no diagram id in document"),
			func(t *testing.T) string {
				return ""
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := zap.NewNop()
			ctrlr := NewController(logger)

			_, err := ctrlr.XmlToItems(context.Background(), logger, bytes.NewReader(test.document), nil)
			assert.Error(t, err)
			assert.Equal(t, test.expectedError, err)

		})
	}
}

func TestController_XmlToItems(t *testing.T) {
	tests := []struct {
		name           string
		document       []byte
		expectedResult int
		expectedError  error
	}{
		{
			// why the hell tests give 6
			"found 7 elements",
			[]byte(`
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
		`),
			7,
			nil,
		},
		{
			"found 0 elements",
			[]byte(`
			<mxfile host="65bd71144e">
				<diagram id="uweCVhkyVy6MirBnUyNJ" name="Page-1">
				<mxGraphModel dx="354" dy="159" grid="1" gridSize="10" guides="1" tooltips="1" connect="0" arrows="1" fold="1" page="1" pageScale="1" pageWidth="827" pageHeight="1169" math="0" shadow="0">
					<root>
					</root>
				</mxGraphModel>
			</diagram>
		</mxfile>
		`),
			0,
			nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := zap.NewNop()
			ctrlr := NewController(logger)
			itemsFound := 0
			chanItems := make(chan Item)

			go func(ch <-chan Item) {
				for {
					select {
					case di := <-ch:
						itemsFound++
						t.Logf("elements found %v, id %v", itemsFound, di.EID)
					default:

					}
				}
			}(chanItems)

			// bellow does not capture the last element
			//go func(ch <-chan diagram.Item) {
			//
			//	for range ch {
			//		itemsFound++
			//		t.Logf("elements found %v", itemsFound)
			//	}
			//}(chanItems)

			_, _ = ctrlr.XmlToItems(context.Background(), logger, bytes.NewReader(test.document), chanItems)
			assert.Equal(t, itemsFound, test.expectedResult)

		})
	}
}
