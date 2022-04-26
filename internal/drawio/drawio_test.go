package drawio

import (
	"encoding/xml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestXMLBasic(t *testing.T) {
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

	tests := []struct {
		name string
	}{
		{
			"import xml",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			//logger := zap.NewNop()
			D := &Mxfile{}
			err := xml.Unmarshal(diagram, D)
			assert.NoError(t, err)
		})
	}
}
