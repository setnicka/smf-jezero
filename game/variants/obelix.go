package variants

import (
	"fmt"
	"html"
	"strings"
)

// Obelix is variant for the Jezero game inspired by Discworld
type Obelix struct{}

// NewObelix creates a new instance of SmallGods
func NewObelix() *Obelix { return &Obelix{} }

////////////////////////////////////////////////////////////////////////////////
//revive:disable:exported

func (h *Obelix) ViewTemplateName() string    { return "viewCarcassonne" }
func (h *Obelix) TemplateStateName() string   { return "Síla čaje" }
func (h *Obelix) TemplateStateSymbol() string { return "" }
func (h *Obelix) TemplateMoneyName() string   { return "Pohoda" }
func (h *Obelix) TemplateMoneySymbol() string { return "" }

func (h *Obelix) NopName() string        { return "Nedělat nic" }
func (h *Obelix) EcoName() string        { return "Poklidný čajový obřad" }
func (h *Obelix) HarvestName() string    { return "Bouřlivé čajové slavnosti" }
func (h *Obelix) CleaningName() string   { return "Vhození lístku čaje" }
func (h *Obelix) InspectionName() string { return "Rozsévání negativní karmy" }
func (h *Obelix) EspionageName() string  { return "Věštba z čajových lístků" }

////////////////////////////////////////////////////////////////////////////////

func (h *Obelix) EcoMessage(money int, pollution int) string {
	return fmt.Sprintf("Popíjeli jste poklidně čaj se svými přáteli. Získali jste tím %d pohody a čaj zeslábl o %d.", money, pollution)
}
func (h *Obelix) HarvestPenaltyMessage(penalty int) string {
	return fmt.Sprintf("Vaše čajové orgie pobouřili ostatní Tibeťany! Vaše karma se v důsledku sociálního nátlaku snížila a ztratili jste %d pohody.", penalty)
}
func (h *Obelix) HarvestSuccessMessage(money int, pollution int) string {
	return fmt.Sprintf("Vaše čajové orgie měly opravodý úspěch. Na party jste se uvolnili a vaše pohoda stoupla o %d. Čaj ale zeslábl o %d.", money, pollution)
}
func (h *Obelix) CleaningMessage(cleaning int) string {
	return fmt.Sprintf("Vhodili jste do kotle několik čajových lístků a jeho barva zmedovatěla. Síla čaje vzrostla o %d.", cleaning)
}
func (h *Obelix) InspectionMessage() string {
	return "Požádali jste o bohy o ochranu čaje. Pokud někdo v minulém kole prováděl něco špatného, tak byl zasažen negativní karmou."
}
func (h *Obelix) EspionageFailMessage() string {
	return "Bohužel čajové lístky byly již příliš staré a pro věštbu zcela nevhodné."
}
func (h *Obelix) EspionageSuccessMessage(teamActions map[string]string) string {
	results := []string{}
	for team, action := range teamActions {
		results = append(results, fmt.Sprintf("<li>%s: <b>%s</b></li>", html.EscapeString(team), action))
	}
	return fmt.Sprintf("Úspěšně jste vyvěštili z čajových lístků. Akce v proběhlém kole:<ul>\n%s\n</ul>", strings.Join(results, "\n"))
}

func (h *Obelix) GlobalMessage(reduce string, increase string, change int) string {
	return fmt.Sprintf("<b>Přelevání:</b> Přeláváním čaje se jeho síla v %s zmenšila o %d a síla v %s se naopak zvedla o %d.", reduce, change, increase, change)
}
